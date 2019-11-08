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

# Takes an old datadir and echoes the expected new datadir path.
upgrade_datadir() {
    local base="$(basename $1)"
    local dir="$(dirname $1)_upgrade"

    # Sanity check.
    [ -n "$base" ]
    [ -n "$dir" ]

    echo "$dir/$base"
}


# yes, this will fail once we allow an index on a partition table
@test "pg_upgrade --check fails on a source cluster with an index on a partition table" {
    skip_if_no_gpdb

    #TODO: code factor this with execute.bats
    PSQL="$GPHOME"/bin/psql

    # Store the data directories for each source segment by port.
    run $PSQL -AtF$'\t' -p $PGPORT postgres -c "select port, datadir from gp_segment_configuration where role = 'p'"
    [ "$status" -eq 0 ] || fail "$output"

    declare -a olddirs
    while read -r port dir; do
        olddirs[$port]="$dir"
    done <<< "$output"

    local masterdir="${olddirs[$PGPORT]}"
    local newmasterdir="$(upgrade_datadir $masterdir)"

    # add in a index on a partition table, which causes pg_upgrade --check to fail
    $PSQL -d postgres -p 6000 -c "create table test_pg_upgrade(a int) distributed by (a) partition by range (a)(start (1) end(4) every(1));"
    $PSQL -d postgres -p 6000 -c "create unique index fomo on test_pg_upgrade (a);"

    gpupgrade initialize \
        --old-bindir "$GPHOME/bin" \
        --new-bindir "$GPHOME/bin" \
        --old-port "$PGPORT" 3>&-

    NEW_CLUSTER="$newmasterdir"

    [ $(grep -c "Checking for indexes on partitioned tables                  fatal" "$GPUPGRADE_HOME"/initialize.log) ] || fail "error expected file: $GPUPGRADE_HOME/initialize.log to contain index-based failure"

    # revert added index
    gpstart -a
    $PSQL -d postgres -p 6000 -c "DROP TABLE test_pg_upgrade CASCADE;"
    gpstop -a

}

@test "gpupgrade initialize runs pg_upgrade --check on master and primaries" {
   skip_if_no_gpdb

  gpupgrade initialize \
      --old-bindir "$GPHOME/bin" \
      --new-bindir "$GPHOME/bin" \
      --old-port "$PGPORT" 3>&-

  # TODO: validate that the master pg_upgrade --check worked too

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




