#!/bin/bash

set -eux -o pipefail

dump_sql() {
    local port=$1
    local dumpfile=$2

    echo "Dumping cluster contents from port ${port} to ${dumpfile}..."

    ssh -n mdw "
        source ${GPHOME_NEW}/greenplum_path.sh
        pg_dumpall -p ${port} -f '$dumpfile'
    "
}

compare_dumps() {
    local old_dump=$1
    local new_dump=$2

    echo "Comparing dumps at ${old_dump} and ${new_dump}..."

    # 5 to 6 requires some massaging of the diff due to expected changes.
    if (( $FILTER_DIFF )); then
        go build ./ci/scripts/filter
        scp ./filter mdw:/tmp/filter

        # First filter out any algorithmically-fixable differences, then
        # patch out the remaining expected diffs explicitly.
        ssh mdw "
            /tmp/filter < '$new_dump' > '$new_dump.filtered'
            patch -R '$new_dump.filtered'
        " < ./ci/scripts/filter/acceptable_diff

        new_dump="$new_dump.filtered"
    fi

    ssh -n mdw "
        diff -U3 --speed-large-files --ignore-space-change --ignore-blank-lines '$old_dump' '$new_dump'
    "
}

# Retrieves the installed GPHOME for a given GPDB RPM.
rpm_gphome() {
    local package_name=$1

    local version=$(ssh -n gpadmin@mdw rpm -q --qf '%{version}' "$package_name")
    echo /usr/local/greenplum-db-$version
}

run_mdw() {
    run_on_host mdw "${1}"
}

run_on_host() {
    local host=$1
    local CMD=$2

    ssh -n $host "
        source ${GPHOME_NEW}/greenplum_path.sh
        ${CMD}
    "
}

check_mirrors() {
    check_segments_are_synchronized
    check_mirror_replication_connections
}

check_segments_are_synchronized() {
    for i in {1..10}; do
        run_mdw "psql -p $MASTER_PORT -d postgres -c \"SELECT gp_request_fts_probe_scan();\""
        local unsynced=$(run_mdw "psql -p $MASTER_PORT -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE content <> -1 AND mode = 'n'\"")
        if [ "$unsynced" = "0" ]; then
            return 0
        fi
        sleep 5
    done

    echo "failed to synchronize within time limit"
    return 1
}

check_mirror_replication_connections() {
    local rows=$(run_mdw "psql -p $MASTER_PORT -d postgres -t -A -c \"select primaries.address, primaries.port, mirrors.hostname FROM
    gp_segment_configuration AS primaries JOIN
    gp_segment_configuration AS mirrors ON
    primaries.content = mirrors.content WHERE
    primaries.role = 'p' AND mirrors.role = 'm' AND primaries.content != -1;\"")
    for row in "${rows[@]}"; do
        local primary_address=$(echo $row | awk '{split($0,a,"|"); print a[1]}')
        local primary_port=$(echo $row | awk '{split($0,a,"|"); print a[2]}')
        local mirror_host=$(echo $row | awk '{split($0,a,"|"); print a[3]}')
        check_replication_connection $primary_address $primary_port $mirror_host
    done
}
check_replication_connection() {
    local primary_address=$1
    local primary_port=$2
    local mirror_host=$3

    local cmd="PGOPTIONS=\"-c gp_session_role=utility\" psql -h $primary_address -p $primary_port  \"dbname=postgres replication=database\" -c \"IDENTIFY_SYSTEM;\""
    run_on_host $mirror_host "$cmd"
}

kill_primaries() {
    local primary_data_dirs=$(run_mdw "psql -p $MASTER_PORT -t -A -d postgres -c \"SELECT hostname, port, datadir FROM gp_segment_configuration WHERE content <> -1 AND role = 'p'\"")
    for pair in ${primary_data_dirs[@]}; do
        local host=$(echo $pair | awk '{split($0,a,"|"); print a[1]}')
        local port=$(echo $pair | awk '{split($0,a,"|"); print a[2]}')
        local dir=$(echo $pair | awk '{split($0,a,"|"); print a[3]}')
        run_on_host $host "pg_ctl stop -p $port -m fast -D $dir -w"
    done
}

wait_can_start_transactions() {
    local host=$1
    local port=$2
    for i in {1..10}; do
        run_on_host $host "psql -p $port -d postgres -c \"SELECT gp_request_fts_probe_scan();\""
        run_on_host $host "psql -p $port -t -A -d postgres -c \"BEGIN; CREATE TEMP TABLE temp_test(a int) DISTRIBUTED RANDOMLY; COMMIT;\""
        if [[ $? -eq 0 ]]; then
            return 0
        fi
        sleep 5
    done

    echo "failed to start transactions within time limit"
    return 1
}

create_table_with_name() {
    local table_name=$1
    local size=$2
    run_mdw "psql -q -p $MASTER_PORT -d postgres -c \"CREATE TABLE ${table_name} (a int) DISTRIBUTED BY (a);\""
    run_mdw "psql -q -p $MASTER_PORT -d postgres -c \"INSERT INTO ${table_name} SELECT * FROM generate_series(0,${size});\""
    get_data_distribution $table_name
}

get_data_distribution() {
    local table_name=$1
    run_mdw "psql -t -A -p $MASTER_PORT -d postgres -c \"SELECT gp_segment_id,count(*) FROM ${table_name} GROUP BY gp_segment_id ORDER BY gp_segment_id;\""
}

check_data_matches() {
    local table_name=$1
    local expected=$2

    local actual=$(get_data_distribution $table_name)
    if [ "${actual}" != "${expected}" ]; then
        echo "Checking table ${table_name} - got: ${actual} want: ${expected}"
        return 1
    fi
}

# Check the validity of the upgraded mirrors - failover to them and then recover, similar to cross-subnet testing
# |  step  |   mdw       | smdw         | sdw-primaries | sdw-mirrors |
# |    1   |   master    |   standby    |    primary    |  mirror     |
# |    2   |   master    |   standby    |      -        |  mirror     |
# |    3   |   master    |   standby    |      -        |  primary    |
# |    4   |   master    |   standby    |   mirror      |  primary    |
# |    5   |   master    |   standby    |   primary     |  mirror     |
check_mirror_validity() {
    local master_data=$(run_mdw "psql -p $MASTER_PORT -t -A -d postgres -c \"SELECT hostname, port, datadir FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_host=$(echo $master_data | awk '{split($0,a,"|"); print a[1]}')
    local master_port=$(echo $master_data | awk '{split($0,a,"|"); print a[2]}')
    local master_data_dir=$(echo $master_data | awk '{split($0,a,"|"); print a[3]}')

    # step 1
    wait_can_start_transactions $master_host $master_port
    check_mirrors

    local on_upgraded_master=$(create_table_with_name on_upgraded_master 50)

    # step 2
    kill_primaries

    # step 3
    wait_can_start_transactions $master_host $master_port

    check_data_matches on_upgraded_master "${on_upgraded_master}"
    local on_promoted_mirrors=$(create_table_with_name on_promoted_mirrors 60)

    # step 4
    run_mdw "export MASTER_DATA_DIRECTORY=${master_data_dir}; export PGPORT=$MASTER_PORT; gprecoverseg -a"  #TODO..why is PGPORT not actually needed here?
    check_mirrors

    check_data_matches on_upgraded_master "${on_upgraded_master}"
    check_data_matches on_promoted_mirrors "${on_promoted_mirrors}"
    local on_recovered_cluster=$(create_table_with_name on_recovered_cluster 70)

    # step 5
    run_mdw "export MASTER_DATA_DIRECTORY=${master_data_dir}; export PGPORT=$MASTER_PORT; gprecoverseg -ra"
    check_mirrors

    check_data_matches on_upgraded_master "${on_upgraded_master}"
    check_data_matches on_promoted_mirrors "${on_promoted_mirrors}"
    check_data_matches on_recovered_cluster "${on_recovered_cluster}"
}

#
# MAIN
#

# This port is selected by our CI pipeline
MASTER_PORT=5432

# We'll need this to transfer our built binaries over to the cluster hosts.
./ccp_src/scripts/setup_ssh_to_cluster.sh

# Cache our list of hosts to loop over below.
mapfile -t hosts < cluster_env_files/hostfile_all

# Copy over the SQL dump we pulled from master.
scp sqldump/dump.sql.xz gpadmin@mdw:/tmp/

# Figure out where GPHOMEs are.
export GPHOME_OLD=$(rpm_gphome ${OLD_PACKAGE})
export GPHOME_NEW=$(rpm_gphome ${NEW_PACKAGE})

# Build gpupgrade.
export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

cd $GOPATH/src/github.com/greenplum-db/gpupgrade
make depend
make

# Install gpupgrade binary onto the cluster machines.
for host in "${hosts[@]}"; do
    scp gpupgrade "gpadmin@$host:/tmp"
    ssh centos@$host "sudo mv /tmp/gpupgrade /usr/local/bin"
done

echo 'Loading SQL dump into source cluster...'
time ssh mdw bash <<EOF
    set -eux -o pipefail

    source ${GPHOME_OLD}/greenplum_path.sh
    export PGOPTIONS='--client-min-messages=warning'
    unxz < /tmp/dump.sql.xz | psql -f - postgres
EOF

# Dump the old cluster for later comparison.
dump_sql $MASTER_PORT /tmp/old.sql

# Now do the upgrade.
time ssh mdw bash <<EOF
    set -eux -o pipefail

    gpupgrade initialize \
              --target-bindir ${GPHOME_NEW}/bin \
              --source-bindir ${GPHOME_OLD}/bin \
              --source-master-port $MASTER_PORT

    gpupgrade execute
    gpupgrade finalize
EOF

# TODO: how do we know the cluster upgraded?  5 to 6 is a version check; 6 to 6 ?????
#   currently, it's sleight of hand...old is on port $MASTER_PORT then new is!!!!

# Dump the new cluster and compare.
dump_sql $MASTER_PORT /tmp/new.sql
if ! compare_dumps /tmp/old.sql /tmp/new.sql; then
    echo 'error: before and after dumps differ'
    exit 1
fi

# Test that mirrors actually work
echo 'Doing failover tests of mirrors...'
check_mirror_validity

echo 'Upgrade successful.'
