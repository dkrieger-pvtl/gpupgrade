#! /usr/bin/env bats

load helpers

setup() {
    skip_if_no_gpdb

    STATE_DIR=`mktemp -d /tmp/gpupgrade.XXXXXX`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"
    gpupgrade kill-services

    # If this variable is set (to a master data directory), teardown() will call
    # gpdeletesystem on this cluster.
    NEW_CLUSTER=
}

teardown() {
    skip_if_no_gpdb

    gpupgrade kill-services
    rm -r "$STATE_DIR"

    if [ -n "$NEW_CLUSTER" ]; then
        delete_cluster $NEW_CLUSTER
    fi

    gpstart -a
}

@test "gpupgrade execute runs gpinitsystem based on the source cluster" {
    skip_if_no_gpdb

    gpupgrade initialize \
        --old-bindir "$GPHOME/bin" \
        --new-bindir "$GPHOME/bin" \
        --old-port "$PGPORT" 3>&-

    [ -s "$STATE_DIR"/pg_upgrade_check_stdout_seg_0.log ] || exit 'error expected file: ...stdout_seg_0.log does not exist'
    [ -r "$STATE_DIR"/pg_upgrade_check_stderr_seg_0.log ] || exit 'error expected file: ...stderr_seg_0.log does not exist'
    [ -s "$STATE_DIR"/pg_upgrade_check_stdout_seg_1.log ] || exit 'error expected file: ...stdout_seg_1.log does not exist'
    [ -r "$STATE_DIR"/pg_upgrade_check_stderr_seg_1.log ] || exit 'error expected file: ...stderr_seg_1.log does not exist'
    [ -s "$STATE_DIR"/pg_upgrade_check_stdout_seg_2.log ] || exit 'error expected file: ...stdout_seg_2.log does not exist'
    [ -r "$STATE_DIR"/pg_upgrade_check_stderr_seg_2.log ] || exit 'error expected file: ...stderr_seg_2.log does not exist'
}
