#! /usr/bin/env bats

load helpers

setup() {
    skip_if_no_gpdb

    STATE_DIR=`mktemp -d`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"
    gpupgrade kill-services

    # If this variable is set (to a master data directory), teardown() will call
    # gpdeletesystem on this cluster.
    NEW_CLUSTER=

    # Store the ports in use on the cluster.
    OLD_PORTS=$(get_ports)

    # Set up an upgrade based on the live cluster, then stop the cluster (to
    # mimic an actual upgrade).
    gpupgrade initialize \
        --old-bindir="$GPHOME/bin" \
        --new-bindir="$GPHOME/bin" \
        --old-port=$PGPORT  \
        --stop-before-cluster-creation \
        --disk-free-ratio 0 3>&-
    gpstop -a
}

teardown() {
    # XXX Beware, BATS_TEST_SKIPPED is not a documented export.
    if [ -z "${BATS_TEST_SKIPPED}" ]; then
        gpupgrade kill-services

        if [ -n "$NEW_CLUSTER" ]; then
            delete_cluster $NEW_CLUSTER
        fi
        rm -rf "$STATE_DIR/demoDataDir*"
        rm -r "$STATE_DIR"

        start_source_cluster
    fi
}

@test "finalize modifies data directories ports on the live target cluster" {

    # To avoid spinning up an entire upgrade just to test finalize, we instead
    # create a new cluster for the test and fake the configurations to point at
    # it.
    #
    # XXX we assume three primaries (demo cluster layout)
    # XXX we hardcode ports here, so we'll fail if there are any conflicts.
    mkdir "$STATE_DIR/_upgrade"
    echo localhost > "$STATE_DIR/hostfile"
    cat - > "$STATE_DIR/gpinitsystem_config" <<EOF
ARRAY_NAME="gpupgrade test cluster"
MASTER_HOSTNAME=localhost
MACHINE_LIST_FILE="$STATE_DIR/hostfile"

MASTER_PORT=40000
PORT_BASE=50000

SEG_PREFIX=demoDataDir
MASTER_DIRECTORY="$STATE_DIR/_upgrade"
declare -a DATA_DIRECTORY=("$STATE_DIR/_upgrade" "$STATE_DIR/_upgrade" "$STATE_DIR/_upgrade")

TRUSTED_SHELL=ssh
CHECK_POINT_SEGMENTS=8
ENCODING=UNICODE
EOF

    # XXX There are always warnings, so ignore them...
    gpinitsystem -ac "$STATE_DIR/gpinitsystem_config" 3>&- || true
    NEW_CLUSTER="$STATE_DIR/_upgrade/demoDataDir-1"

    # Mimic the old cluster datadirs, which relies on the above hardcoded
    # gpinitsytem_config
    OLD_DATADIRS="${STATE_DIR}/_/demoDataDir-1
${STATE_DIR}/_/demoDataDir0
${STATE_DIR}/_/demoDataDir1
${STATE_DIR}/_/demoDataDir2"

    while IFS= read -r datadir; do
        echo "mkdir -p $datadir"
    done <<< "$OLD_DATADIRS"

    # Create a marker file for testing to verify old and new clusters actually
    # got reconfigured.
    touch ${STATE_DIR}/_/source.cluster

    # Generate a new target cluster configuration that the hub can use, then
    # restart the hub.
    PGPORT=40000 go run ./testutils/insert_target_config "$GPHOME/bin" "$GPUPGRADE_HOME/config.json"
    gpupgrade kill-services
    gpupgrade hub --daemonize 3>&-

    gpupgrade finalize
    # Reset NEW_CLUSTER for cleanup since finalize reconfigures the datadirs.
    NEW_CLUSTER="$STATE_DIR/_/demoDataDir-1"

    # Check to make sure the new cluster's ports match the old one.
    local new_ports=$(get_ports)
    [ "$OLD_PORTS" = "$new_ports" ] || fail "actual ports: $new_ports"

    # Ensure the new cluster's data dirs match the old one.
    local new_datadirs=$(get_datadirs)
    [ "$OLD_DATADIRS" = "$new_datadirs" ] || fail "actual datadirs: $new_datadirs, expected datadirs: $OLD_DATADIRS"

    [ -f ${STATE_DIR}/_old/source.cluster ] || fail "expected source.cluster marker file to be in source datadir: ${STATE_DIR}/_old"
    [ ! -f ${STATE_DIR}/_/source.cluster ] || fail "unexpecetd source.cluster marker file in target datadir: ${STATE_DIR}/_"
}

# Writes the primary ports from the cluster pointed to by $PGPORT to stdout, one
# per line, sorted by content ID.
get_ports() {
    PSQL="$GPHOME"/bin/psql
    $PSQL -At postgres \
        -c "select port from gp_segment_configuration where role = 'p' order by content"
}

# Writes the datadirs from the cluster pointed to by $PGPORT to stdout, one per
# line, sorted by content ID.
get_datadirs() {
    PSQL="$GPHOME"/bin/psql
    local version=$("$GPHOME"/bin/postgres --gp-version)
    local prefix="postgres (Greenplum Database) "

    if [[ $version == ${prefix}"5"* ]]; then
         $PSQL -At postgres \
                -c "SELECT fselocation
        FROM pg_catalog.gp_segment_configuration
        JOIN pg_catalog.pg_filespace_entry on (dbid = fsedbid)
        JOIN pg_catalog.pg_filespace fs on (fsefsoid = fs.oid)
        ORDER BY content DESC, fs.oid;"
    else
        $PSQL -At postgres \
            -c "select datadir from gp_segment_configuration where role = 'p' order by content"
    fi
}
