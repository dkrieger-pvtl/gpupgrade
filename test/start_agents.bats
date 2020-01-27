#! /usr/bin/env bats

load helpers

setup() {
    skip_if_no_gpdb

    AGENT_PORT=6416
    HELD_PORT_PID=

    STATE_DIR=`mktemp -d`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"

    #TODO: with this and the matching one in teardown() this test hangs
    gpupgrade kill-services
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
    fi
}


hold_onto_port_for_agent() {
    echo "holding onto port $AGENT_PORT"
    nc -l ::0 $AGENT_PORT &
    HELD_PORT_PID=$!
}
initialize_should_be_success() {
    [ "$status" -eq 0 ] || fail "expected start_agent substep to succeed with no other process on its port: $output"
}
initialize_should_be_failure() {
    [ "$status" -ne 0 ] || fail "expected start_agent substep to fail with another process on its port: $output"
}
release_port() {
    echo "releasing: port='$AGENT_PORT' pid='$HELD_PORT_PID'"
    kill $HELD_PORT_PID
}
run_gpupgrade_initialize() {
    run gpupgrade initialize \
                 --old-bindir="$GPHOME/bin" \
                 --new-bindir="$GPHOME/bin" \
                 --old-port=$PGPORT  \
                 --disk-free-ratio 0 3>&-
}


@test "start_agents fails if a process is connected on the same TCP port" {

    hold_onto_port_for_agent
    run_gpupgrade_initialize
    initialize_should_be_failure

    release_port
    run_gpupgrade_initialize
    initialize_should_be_success

}

