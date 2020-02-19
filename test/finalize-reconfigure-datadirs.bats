#!/usr/bin/env bats

load helpers

setup() {
    skip_if_no_gpdb

    [ ! -z $GPHOME ]
    GPHOME_NEW=${GPHOME_NEW:-$GPHOME}
    GPHOME_OLD=$GPHOME

    setup_state_dir

    gpupgrade kill-services
}

teardown() {
    print_teardown_banner
    teardown_new_cluster
    gpupgrade kill-services

    # reload old path and start
    source "${GPHOME_OLD}/greenplum_path.sh"
    gpstart -a
}

@test "it swaps out the target cluster's data directories and archives the source cluster's data directories" {
    log "Using state directory: $STATE_DIR"

    place_marker_file_in_source_cluster

    log "initialize"
    gpupgrade initialize \
        --old-bindir="$GPHOME/bin" \
        --new-bindir="$GPHOME_NEW/bin" \
        --old-port="${PGPORT}" \
        --disk-free-ratio 0 \
        --verbose

    log "execute"
    gpupgrade execute --verbose

    log "finalize"
    gpupgrade finalize

    local source_cluster_master_data_directory="${MASTER_DATA_DIRECTORY}_old"
    local target_cluster_master_data_directory="${MASTER_DATA_DIRECTORY}"

    [ -f "${source_cluster_master_data_directory}/source-cluster.test-marker" ] || fail "expected source-cluster.test-marker marker file to be in source datadir: ${STATE_DIR}/base/demoDataDir-1"
    [ ! -f "${target_cluster_master_data_directory}/source-cluster.test-marker" ] || fail "unexpected source-cluster.test-marker marker file in target datadir: ${STATE_DIR}/base/demoDataDir-1"

    local gpperfmon_config_file="${target_cluster_master_data_directory}/gpperfmon/conf/gpperfmon.conf"

    grep "${target_cluster_master_data_directory}" "${gpperfmon_config_file}" || \
        fail "got gpperfmon.conf file $(cat $gpperfmon_config_file), wanted it to include ${target_cluster_master_data_directory}"

    # [x] TODO: ensure upgrading from 5x works
    # [x] TODO: gpperfmon?
    # [ ] TODO: push segment work to the agent
    # [ ] TODO: ensure link mode works
    # [ ] TODO: ensure new cluster is queryable
    #
    # [-] TODO: ensure old cluster can still start (punting for now)
}


place_marker_file_in_source_cluster() {
    touch "$MASTER_DATA_DIRECTORY/source-cluster.test-marker"
}

setup_state_dir() {
    STATE_DIR=$(mktemp -d /tmp/gpupgrade.XXXXXX)
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"
}

teardown_new_cluster() {
    delete_finalized_cluster $MASTER_DATA_DIRECTORY
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
