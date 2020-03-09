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

check_segments_are_synchronized() {
    for i in {1..10}; do
        run_mdw "psql -p 5432 -d postgres -c \"SELECT gp_request_fts_probe_scan();\""
        local unsyncedMirrors=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE content <> -1 AND mode = 'n'\"")
        if [ "$unsyncedMirrors" = "0" ]; then
            return 0
        fi
        sleep 5
    done

    echo "failed to synchronize within time limit"
    return 1
}

has_standby() {
    local hasStandby=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE content = -1 AND role = 'm'\"")
    if [ "$hasStandby" = "1" ]; then
        return 0
    fi

    echo "cluster has no standby"
    return 1
}

check_standby_is_synchronized() {
    local host=$1
    local port=$2

    for i in {1..10}; do
        run_on_host $host "psql -p $port -d postgres -c \"SELECT gp_request_fts_probe_scan();\""
        local unsynchedStandby=$(run_mdw "psql -p $port -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE content = -1 AND mode = 'n' AND role = 'm'\"")
        if [ "$unsynchedStandby" = "0" ]; then
            return 0
        fi
        sleep 5
    done

    echo "failed to synchronize within time limit"
    return 1
}

kill_primaries() {
    local primary_data_dirs=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT hostname, port, datadir FROM gp_segment_configuration WHERE content <> -1 AND role = 'p'\"")
    for pair in ${primary_data_dirs[@]}; do
        local host=$(echo $pair | awk '{split($0,a,"|"); print a[1]}')
        local port=$(echo $pair | awk '{split($0,a,"|"); print a[2]}')
        local dir=$(echo $pair | awk '{split($0,a,"|"); print a[3]}')
        run_on_host $host "pg_ctl stop -p $port -m fast -D $dir -w"
    done
}

kill_master_on_host() {
    local host=$1
    local port=$2
    local master_data_dir=$(run_on_host $host "psql -p $port -t -A -d postgres -c \"SELECT datadir FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    run_on_host $host "pg_ctl stop -p $port -m fast -D $master_data_dir -w"
}

activate_standby() {
    local host=$1
    local port=$2
    local datadir=$3
    run_on_host $host "export PGPORT=$port; gpactivatestandby -a -d $datadir"
}

check_can_start_transactions() {
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

check_synchronized_after_standby_change() {
    local host=$1
    local port=$2
    for i in {1..20}; do
        # todo: why does running an fts probe or committing a transaction hang?
        local unsyncedMirrors=$(run_on_host $host "psql -p $port -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE role = 'm' AND mode = 'n'\"")
        if [ "$unsyncedMirrors" = "0" ]; then
            return 0
        fi
        sleep 6
    done

    echo "failed to synchronize within time limit after changing standby"
    return 1
}

check_standby_synchronized_after_initialize() {
    local host=$1
    local port=$2
    for i in {1..20}; do
        # todo: why does running an fts probe or committing a transaction hang?
        local syncedStandby=$(run_on_host $host "psql -p $port -t -A -d postgres -c \"SELECT count(*) FROM gp_segment_configuration WHERE role = 'm' AND content = -1 AND mode = 's'\"")
        if [ "$syncedStandby" = "1" ]; then
            return 0
        fi
        sleep 6
    done

    echo "failed to synchronize within time limit after initializing standby"
    return 1
}

check_replication_connection() {
    local primary_address=$1
    local primary_port=$2
    local mirror_host=$3

    run_on_host $mirror_host PGOPTIONS="-c gp_session_role=utility" psql -h $primary_address -p $primary_port  "dbname=postgres replication=database" -c "IDENTIFY_SYSTEM;"
}

get_data_distribution() {
    local table_name=$1
    run_mdw "psql -t -A -p 5432 -d postgres -c \"SELECT gp_segment_id,count(*) FROM ${table_name} GROUP BY gp_segment_id ORDER BY gp_segment_id;\""
}

create_table_with_name() {
    local table_name=$1
    local size=$2
    run_mdw "psql -q -p 5432 -d postgres -c \"CREATE TABLE ${table_name} (a int) DISTRIBUTED BY (a);\""
    run_mdw "psql -q -p 5432 -d postgres -c \"INSERT INTO ${table_name} SELECT * FROM generate_series(0,${size});\""
    get_data_distribution $table_name
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

move_data_dir() {
    local host=$1
    local src=$2
    run_on_host $1 "mv $src $(mktemp -d)"
}

initialize_standby() {
    local master_host=$1
    local master_port=$2
    local standby_host=$3
    local port=$4
    local datadir=$5

    run_on_host $master_host "export PGPORT=$master_port; gpinitstandby -a -s $standby_host -P $port -S $datadir"
}

# Check the validity of the upgraded mirrors - failover to them and then recover, similar to cross-subnet testing
# |  step  |   mdw       | smdw         | sdw-primaries | sdw-mirrors |
# |    1   |   master    |   standby    |    primary    |  mirror     |
# |    2   |   master    |   standby    |      -        |  mirror     |
# |    3   |   master    |   standby    |      -        |  primary    |
# |    4   |   master    |   standby    |   mirror      |  primary    |
# |    5   |   master    |   standby    |   primary     |  mirror     |
check_mirror_validity() {
    local master_host=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT hostname FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_port=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT port FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_data_dir=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT datadir FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")

    # add a table that is distributed across the cluster (data1)
    local on_upgraded_master=$(create_table_with_name on_upgraded_master 50)

    # step 1
    check_can_start_transactions $master_host $master_port
    check_segments_are_synchronized

    # step 2
    kill_primaries

    # step 3
    check_can_start_transactions $master_host $master_port

    # check that (data1) is on the failed over mirrors that are acting as primaries
    check_data_matches on_upgraded_master "${on_upgraded_master}"
    local on_promoted_mirrors=$(create_table_with_name on_promoted_mirrors 60)

    # step 4
    run_mdw "export MASTER_DATA_DIRECTORY=${master_data_dir}; export PGPORT=5432; gprecoverseg -a"  #TODO..why is PGPORT not actually needed here?
    check_segments_are_synchronized

    # check data1 and data2
    check_data_matches on_upgraded_master "${on_upgraded_master}"
    check_data_matches on_promoted_mirrors "${on_promoted_mirrors}"
    local on_recovered_cluster=$(create_table_with_name on_recovered_cluster 70)

    # step 5
    run_mdw "export MASTER_DATA_DIRECTORY=${master_data_dir}; export PGPORT=5432; gprecoverseg -ra"
    check_segments_are_synchronized

    # check data1,2,3
    check_data_matches on_upgraded_master "${on_upgraded_master}"
    check_data_matches on_promoted_mirrors "${on_promoted_mirrors}"
    check_data_matches on_recovered_cluster "${on_recovered_cluster}"
}

# TODO: add replication connections("replication connections can be made from the acting" from cross_subet.py")

# https://gpdb.docs.pivotal.io/6-4/admin_guide/highavail/topics/g-restoring-master-mirroring-after-a-recovery.html#topic17
# | step  | mdw     | smdw    | sdw-primaries | sdw-mirrors |
# |  1    | master  | standby |    primary    |  mirror     |
# |  2    | -       | standby |    primary    |  mirror     |
# |  3    | -       | master  |    primary    |  mirror     |
# |  4    | standby | master  |    primary    |  mirror     |   move original master somewhere else or delete it
# |  5    | standby | -       |    primary    |  mirror     |
# |  6    | master  | -       |    primary    |  mirror     |
# |  7    | master  | standby |    primary    |  mirror     |   move original standby somewhere else or delete it
check_standby_validity() {
    local master_host=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT hostname FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_address=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT address FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_port=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT port FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")
    local master_data_dir=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT datadir FROM gp_segment_configuration WHERE content = -1 AND role = 'p'\"")

    local standby_host=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT hostname FROM gp_segment_configuration WHERE content = -1 AND role = 'm'\"")
    local standby_address=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT address FROM gp_segment_configuration WHERE content = -1 AND role = 'm'\"")
    local standby_port=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT port FROM gp_segment_configuration WHERE content = -1 AND role = 'm'\"")
    local standby_data_dir=$(run_mdw "psql -p 5432 -t -A -d postgres -c \"SELECT datadir FROM gp_segment_configuration WHERE content = -1 AND role = 'm'\"")

    # step 1
    has_standby
    # make sure master-standby in sync here...

    # step 2
    kill_master_on_host $master_host $master_port

    # step 3
    activate_standby $standby_host $standby_port $standby_data_dir
    check_synchronized_after_standby_change $standby_host $standby_port

    # step 4
    local orig_master_data_dir=$(move_data_dir $master_host $master_data_dir)
    initialize_standby $standby_host $standby_port $master_host $master_port $master_data_dir
    check_standby_synchronized_after_initialize $standby_host $standby_port
    check_synchronized_after_standby_change $standby_host $standby_port

    check_replication_connection $standby_address $standby_port $master_host

    # step 5
    kill_master_on_host $standby_host $standby_port

    # step 6
    activate_standby $master_host $master_port $master_data_dir
    check_synchronized_after_standby_change $master_host $master_port

    # step 7
    local orig_standby_data_dir=$(move_data_dir $standby_host $standby_data_dir)
    initialize_standby $master_host $master_port $standby_host $standby_port $standby_data_dir
    check_standby_synchronized_after_initialize $master_host $master_port
    check_synchronized_after_standby_change $master_host $master_port
    check_replication_connection $master_address $master_port $standby_host
}

#
# MAIN
#

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

echo 'Loading SQL dump into old cluster...'
time ssh mdw bash <<EOF
    set -eux -o pipefail

    source ${GPHOME_OLD}/greenplum_path.sh
    export PGOPTIONS='--client-min-messages=warning'
    unxz < /tmp/dump.sql.xz | psql -f - postgres
EOF

# Dump the old cluster for later comparison.
dump_sql 5432 /tmp/old.sql

# Now do the upgrade.
time ssh mdw bash <<EOF
    set -eux -o pipefail

    gpupgrade initialize \
              --target-bindir ${GPHOME_NEW}/bin \
              --source-bindir ${GPHOME_OLD}/bin \
              --source-master-port 5432

    gpupgrade execute
    gpupgrade finalize
EOF

# TODO: how do we know the cluster upgraded?  5 to 6 is a version check; 6 to 6 ?????
#   currently, it's sleight of hand...old is on port 5432 then new is!!!!

# Dump the new cluster and compare.
dump_sql 5432 /tmp/new.sql
if ! compare_dumps /tmp/old.sql /tmp/new.sql; then
    echo 'error: before and after dumps differ'
    exit 1
fi

# Test that the standby and mirrors actually work
echo 'Doing failover tests of standby and mirrors...'
check_mirror_validity
check_standby_validity

echo 'Upgrade successful.'
