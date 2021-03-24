#!/bin/bash
#
# Copyright (c) 2017-2021 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

set -eux -o pipefail

is_GPDB5() {
    local gphome=$1
    version=$(ssh mdw "$gphome"/bin/postgres --gp-version)
    [[ $version =~ ^"postgres (Greenplum Database) 5." ]]
}

# set the database gucs
# 1. bytea_output: by default for bytea the output format is hex on GPDB 6,
#    so change it to escape to match GPDB 5 representation
configure_gpdb_gucs() {
    local gphome=$1
    ssh mdw bash <<EOF
        set -eux -o pipefail

        source ${gphome}/greenplum_path.sh
        export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1
        gpconfig -c bytea_output -v escape
        gpstop -u
EOF
}

# We call this before reindex_all_dbs() to prevent a very large number of warnings
# of the form: "WARNING:  database with OID 0 must be vacuumed within 147482491 transactions"
#  This might only be necessary because of the large number of tables we are currently dropping.
vacuum_all_dbs() {
    local gphome=$1

    databases=$(ssh -n gpadmin@mdw "
      set -x

      source ${gphome}/greenplum_path.sh
      export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

      psql -d regression --tuples-only --no-align --field-separator ' ' <<SQL_EOF
          SELECT datname
          FROM	pg_database
          WHERE	datname != 'template0' and datname not like 'funny%';
SQL_EOF
")

    echo 'VACUUM all databases'
    echo "${databases}" | while read -r database; do
        if [[ -n "${database}" ]]; then
            ssh -n gpadmin@mdw "
                set -x
                source ${gphome}//greenplum_path.sh
                export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1

                psql -d ${database} <<SQL_EOF
                    VACUUM;
SQL_EOF
            "
        fi
done
}

reindex_all_dbs() {
    local gphome=$1
    ssh mdw bash <<EOF
        set -eux -o pipefail

        source ${gphome}/greenplum_path.sh
        export MASTER_DATA_DIRECTORY=/data/gpdata/master/gpseg-1
        reindexdb -a
EOF
}

dump_sql() {
    local port=$1
    local dumpfile=$2

    echo "Dumping cluster contents from port ${port} to ${dumpfile}..."

    ssh -n mdw "
        source ${GPHOME_TARGET}/greenplum_path.sh
        pg_dumpall -p ${port} -f '$dumpfile'
    "
}

compare_dumps() {
    local source_dump=$1
    local target_dump=$2

    echo "Comparing dumps at ${source_dump} and ${target_dump}..."

    pushd gpupgrade_src
        # 5 to 6 requires some massaging of the diff due to expected changes.
        if (( $FILTER_DIFF )); then
            go build ./ci/scripts/filters/filter
            scp ./filter mdw:/tmp/filter

            # First filter out any algorithmically-fixable differences, then
            # patch out the remaining expected diffs explicitly.
            ssh mdw "
                /tmp/filter -version=6 -inputFile='$target_dump' > '$target_dump.filtered'
                patch -R '$target_dump.filtered'
            " < ./ci/scripts/filters/${DIFF_FILE}

            target_dump="$target_dump.filtered"

            # Run the filter on the source dump
            ssh mdw "
                /tmp/filter -version=5 -inputFile='$source_dump' > '$source_dump.filtered'
            "

            source_dump="$source_dump.filtered"
        fi
    popd

    ssh -n mdw "
        diff -U3 --speed-large-files --ignore-space-change --ignore-blank-lines '$source_dump' '$target_dump'
    "
}

#
# MAIN
#

# Global parameters (default to off)
USE_LINK_MODE=${USE_LINK_MODE:-0}
FILTER_DIFF=${FILTER_DIFF:-0}
DIFF_FILE=${DIFF_FILE:-"icw.diff"}
GPDB_6_TO_7_SQL_DROP_OBJECTS=${GPDB_6_TO_7_SQL_DROP_OBJECTS:-0}

# This port is selected by our CI pipeline
MASTER_PORT=5432

# We'll need this to transfer our built binaries over to the cluster hosts.
./ccp_src/scripts/setup_ssh_to_cluster.sh

# Cache our list of hosts to loop over below.
mapfile -t hosts < cluster_env_files/hostfile_all

export GPHOME_SOURCE=/usr/local/greenplum-db-source
export GPHOME_TARGET=/usr/local/greenplum-db-target

# This script should eventually be
if (( $GPDB_6_TO_7_SQL_DROP_OBJECTS )); then
    echo "GPDB_6_TO_7_SQL_DROP_OBJECTS..."
    scp gpupgrade_src/ci/scripts/load-dump-6to7-drop-objects.bash "gpadmin@mdw:/tmp"
    ssh gpadmin@mdw "/bin/bash /tmp/load-dump-6to7-drop-objects.bash"
fi

for host in "${hosts[@]}"; do
    scp rpm_enterprise/gpupgrade-*.rpm "gpadmin@$host:/tmp"
    ssh centos@$host "sudo rpm -ivh /tmp/gpupgrade-*.rpm"
done

# On GPDB version other than 5, set the gucs before taking dumps
if ! is_GPDB5 ${GPHOME_SOURCE}; then
    configure_gpdb_gucs ${GPHOME_SOURCE}
fi

# SKIPPING FOR NOW....RENABBLE
# Dump the old cluster for later comparison.
#dump_sql $MASTER_PORT /tmp/source.sql

# Now do the upgrade.
LINK_MODE=""
if [ "${USE_LINK_MODE}" = "1" ]; then
    LINK_MODE="--mode=link"
fi

time ssh mdw bash <<EOF
    set -eux -o pipefail

    gpupgrade initialize \
              $LINK_MODE \
              --automatic \
              --target-gphome ${GPHOME_TARGET} \
              --source-gphome ${GPHOME_SOURCE} \
              --source-master-port $MASTER_PORT \
              --temp-port-range 6020-6040 \
              --skip-version-check \
              --verbose
    # TODO: rather than setting a temp port range, consider carving out an
    # ip_local_reserved_ports range during/after CCP provisioning.

    gpupgrade execute --non-interactive --verbose
    gpupgrade finalize --non-interactive --verbose
EOF

# On GPDB version other than 5, set the gucs before taking dumps
# and reindex all the databases to enable bitmap indexes which were
# marked invalid during upgrade
if ! is_GPDB5 ${GPHOME_TARGET}; then
    configure_gpdb_gucs ${GPHOME_TARGET}
    # TODO: vacuum_all_dbs might only be needed on a gpdb-7 target cluster
    # but this fails right now due to a likely server issue....
    # need to fix...
    # vacuum_all_dbs ${GPHOME_TARGET}
    reindex_all_dbs ${GPHOME_TARGET}
fi

# TODO: how do we know the cluster upgraded?  5 to 6 is a version check; 6 to 6 ?????
#   currently, it's sleight of hand...source is on port $MASTER_PORT then target is!!!!
#   perhaps use the controldata("pg_controldata $MASTER_DATA_DIR") system identifier?

# Dump the target cluster and compare.
dump_sql ${MASTER_PORT} /tmp/target.sql
if ! compare_dumps /tmp/source.sql /tmp/target.sql; then
    echo 'error: before and after dumps differ'
    exit 1
fi

echo 'Upgrade successful.'
