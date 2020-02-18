#!/usr/bin/env bats

load helpers

setup() {
    skip_if_no_gpdb

    setup_state_dir

    log $STATE_DIR

    gpupgrade kill-services
}

teardown() {
    teardown_new_cluster
    gpupgrade kill-services
    gpstart -a

    echo "done"
}

@test "it swaps out the target cluster's data directories and archives the source cluster's data directories" {
    place_marker_file_in_source_cluster

    gpupgrade initialize \
        --old-bindir="$GPHOME/bin" \
        --new-bindir="$GPHOME/bin" \
        --old-port="${PGPORT}" \
        --disk-free-ratio 0 \
        --verbose

    gpupgrade execute --verbose

    gpupgrade finalize

    local source_cluster_master_data_directory="${MASTER_DATA_DIRECTORY}_old"
    local target_cluster_master_data_directory="${MASTER_DATA_DIRECTORY}"

    [ -f "${source_cluster_master_data_directory}/source-cluster.test-marker" ] || fail "expected source-cluster.test-marker marker file to be in source datadir: ${STATE_DIR}/base/demoDataDir-1"
    [ ! -f "${target_cluster_master_data_directory}/source-cluster.test-marker" ] || fail "unexpected source-cluster.test-marker marker file in target datadir: ${STATE_DIR}/base/demoDataDir-1"

    # TODO: ensure upgrading from 5x works
    # TODO: gpperfmon?
    # TODO: ensure old cluster can still start
    # TODO: push segment work to the agent
}

place_marker_file_in_source_cluster() {
    touch "$MASTER_DATA_DIRECTORY/source-cluster.test-marker"
}

setup_state_dir() {
    STATE_DIR=$(mktemp -d /tmp/gpupgrade.XXXXXX)
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"
}

teardown_new_cluster() {
    local NEW_CLUSTER="$(gpupgrade config show --new-datadir)"

    if [ -n "$NEW_CLUSTER" ]; then
        delete_finalized_cluster $NEW_CLUSTER
    fi
}


## Writes the datadirs from the cluster pointed to by $PGPORT to stdout, one per
## line, sorted by content ID.
#get_datadirs_from_gp_segment_configuration() {
#    PSQL="$GPHOME"/bin/psql
#    local version=$("$GPHOME"/bin/postgres --gp-version)
#    local prefix="postgres (Greenplum Database) "
#
#    if [[ $version == ${prefix}"5"* ]]; then
#         $PSQL -At postgres \
#                -c "SELECT fselocation
#        FROM pg_catalog.gp_segment_configuration
#        JOIN pg_catalog.pg_filespace_entry on (dbid = fsedbid)
#        JOIN pg_catalog.pg_filespace fs on (fsefsoid = fs.oid)
#        ORDER BY content DESC, fs.oid;"
#    else
#        $PSQL -At postgres \
#            -c "select datadir from gp_segment_configuration where role = 'p' order by content"
#    fi
#}
