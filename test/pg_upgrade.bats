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

@test "gpupgrade initialize runs pg_upgrade --check on master and primaries" {
    skip_if_no_gpdb

    gpupgrade initialize \
        --old-bindir "$GPHOME/bin" \
        --new-bindir "$GPHOME/bin" \
        --old-port "$PGPORT" 3>&-

    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_0.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_0.log does not exist"
    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_1.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_1.log does not exist"
    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_2.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_2.log does not exist"

    [ -s "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_0.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_0.log to have logs"
    [ -s "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_1.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_1.log to have logs"
    [ -s "$GPUPGRADE_HOME"/pg_upgrade_check_stdout_seg_2.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stdout_seg_2.log to have logs"

    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_0.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_0.log does not exist"
    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_1.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_1.log does not exist"
    [ -e "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_2.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_2.log does not exist"

    [ ! -s "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_0.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_0.log to have zero size"
    [ ! -s "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_1.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_1.log to have zero size"
    [ ! -s "$GPUPGRADE_HOME"/pg_upgrade_check_stderr_seg_2.log ] || fail "error expected file: $GPUPGRADE_HOME/pg_upgrade_check_stderr_seg_2.log to have zero size"
}
