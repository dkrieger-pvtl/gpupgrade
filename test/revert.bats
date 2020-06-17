#! /usr/bin/env bats
#
# Copyright (c) 2017-2020 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

load helpers

setup() {
    skip_if_no_gpdb

    STATE_DIR=`mktemp -d /tmp/gpupgrade.XXXXXX`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"
    echo $GPUPGRADE_HOME

    gpupgrade kill-services

    PSQL="$GPHOME_SOURCE"/bin/psql
}


teardown() {
    skip_if_no_gpdb

    if [ -n "$TABLE" ]; then
        $PSQL postgres -c "DROP TABLE ${TABLE}"
    fi
}

@test "reverting after initialize succeeds" {
    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    local target_hosts_dirs=$(jq -r '.Target.Primaries[] | .DataDir' "${GPUPGRADE_HOME}/config.json")

    gpupgrade revert --verbose

    # gpupgrade processes are stopped
    ! process_is_running "[g]pupgrade hub" || fail 'expected hub to have been stopped'
    ! process_is_running "[g]pupgrade agent" || fail 'expected agent to have been stopped'

    # target data directories are deleted
    while read -r datadir; do
        run stat "$datadir"
        ! [ $status -eq 0 ] || fail "expected datadir ${datadir} to have been deleted"
    done <<< "${target_hosts_dirs}"

    # the GPUPGRADE_HOME directory is deleted
    if [ -d "${GPUPGRADE_HOME}" ]; then
        echo "expected GPUPGRADE_HOME directory ${GPUPGRADE_HOME} to have been deleted"
        exit 1
    fi

    # check that the archived log directory was created within the last 3 minutes
    if [[ -z $(find "${HOME}/gpAdminLogs/gpupgrade-"* -type d -cmin -3) ]]; then
        fail "expected the log directory to be archived and match ${HOME}/gpAdminLogs/gpupgrade-*"
    fi
}

@test "reverting after execute in copy mode succeeds" {
    setup_restore_cluster "--mode=copy"

    local old_config=$(get_segment_configuration "${GPHOME_SOURCE}")

    local target_master_port=6020

    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range ${target_master_port}-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    # On GPDB5, restore the primary and master directories before starting the cluster. Hack until revert handles this case
    restore_cluster

    gpupgrade revert --verbose

    # Check to make sure the new cluster matches the old one.
    local new_config=$(get_segment_configuration "${GPHOME_SOURCE}")
    [ "$old_config" = "$new_config" ] || fail "actual config: $new_config, wanted: $old_config"

    isready || fail "expected source cluster to be running on port ${PGPORT}"
    ! isready "${GPHOME_TARGET}" ${target_master_port} || fail "expected target cluster to not be running on port ${target_master_port}"
}

@test "reverting after execute in link mode succeeds" {
    local target_master_port=6020

    # Add a table
    TABLE="should_be_reverted"
    $PSQL postgres -c "CREATE TABLE ${TABLE} (a INT)"
    $PSQL postgres -c "INSERT INTO ${TABLE} VALUES (1), (2), (3)"

    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range ${target_master_port}-6040 \
        --disk-free-ratio 0 \
        --mode link \
        --verbose 3>&-
    gpupgrade execute --verbose

    # Modify the table on the target cluster
    $PSQL -p $target_master_port postgres -c "TRUNCATE ${TABLE}"

    # Revert
    gpupgrade revert --verbose

    # Check that transactions can be started on the source
    $PSQL postgres --single-transaction -c "SELECT version()" || fail "unable to start transaction"

    # Verify the table modifications were reverted
    local row_count=$($PSQL postgres -Atc "SELECT COUNT(*) FROM ${TABLE}")
    if (( row_count != 3 )); then
        fail "table ${TABLE} truncated after execute was not reverted: got $row_count rows want 3"
    fi
}

@test "can successfully run gpupgrade after a revert" {
    setup_restore_cluster "--mode=copy"

    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    # On GPDB5, restore the primary and master directories before starting the cluster. Hack until revert handles this case
    restore_cluster

    gpupgrade revert --verbose

    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    # On GPDB5, restore the primary and master directories before starting the cluster. Hack until revert handles this case
    restore_cluster

    # This last revert is used for test cleanup.
    gpupgrade revert --verbose
}

# TODO: this currently falis as we do not recover the source tablespaces on revert yet
#  error: table batsTable truncated after execute was not reverted: got 0 rows want 3
@test "reverting after execute in link mode succeeds with a tablespace" {
    if ! is_GPDB5 "$GPHOME_SOURCE"; then
      skip "only runs on a GPDB5 source cluster"
    fi

    local target_master_port=6020

   # create a filespace ....
    local FILESPACE_ROOT=`mktemp -d /tmp/gpupgrade.XXXXXX`
    for dir in master primary1 primary2 primary3 mirror1 mirror2 mirror3; do
      mkdir -p ${FILESPACE_ROOT}/${dir}
    done

    local HOSTNAME=`hostname`
    cat <<EOF > "${FILESPACE_ROOT}/filespace.txt"
filespace:batsFS
${HOSTNAME}:1:${FILESPACE_ROOT}/master/demoDataDir-1
${HOSTNAME}:2:${FILESPACE_ROOT}/primary1/demoDataDir0
${HOSTNAME}:3:${FILESPACE_ROOT}/primary2/demoDataDir1
${HOSTNAME}:4:${FILESPACE_ROOT}/primary3/demoDataDir2
${HOSTNAME}:5:${FILESPACE_ROOT}/mirror1/demoDataDir0
${HOSTNAME}:6:${FILESPACE_ROOT}/mirror2/demoDataDir1
${HOSTNAME}:7:${FILESPACE_ROOT}/mirror3/demoDataDir2
${HOSTNAME}:8:${FILESPACE_ROOT}/master/standby
EOF

    psql -d postgres -c "DROP TABLE IF EXISTS batsTable;"
    psql -d postgres -c "DROP TABLESPACE IF EXISTS batsTbsp;"
    psql -d postgres -c "DROP FILESPACE IF EXISTS batsFS;"

    gpfilespace -c ${FILESPACE_ROOT}/filespace.txt

    psql -d postgres -c "CREATE TABLESPACE batsTbsp FILESPACE batsFS;"
    psql -d postgres -c "CREATE TABLE batsTable(a int) TABLESPACE batsTbsp;"
    psql -d postgres -c "INSERT INTO batsTable SELECT i from generate_series(1,5)i;"


    gpupgrade initialize \
        --source-bindir="$GPHOME_SOURCE/bin" \
        --target-bindir="$GPHOME_TARGET/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range ${target_master_port}-6040 \
        --disk-free-ratio 0 \
        --mode link \
        --verbose 3>&-
    gpupgrade execute --verbose

    # Modify the table on the target cluster
    $PSQL -p $target_master_port postgres -c "TRUNCATE batsTable"

    # Revert
    gpupgrade revert --verbose

    # Check that transactions can be started on the source
    $PSQL postgres --single-transaction -c "SELECT version()" || fail "unable to start transaction"

    # Verify the table modifications were reverted
    local row_count=$($PSQL postgres -Atc "SELECT COUNT(*) FROM batsTable")
    if (( row_count != 5 )); then
        fail "table batsTable truncated after execute was not reverted: got $row_count rows want 3"
    fi
}

# TODO This test currently fails as we do not delete the target directory tablespaces
#   Upgrading master...                                                [FAILED]
#   This is on the second execute...as the tablespace exists.  By hand, the error looks like:
# psql:pg_upgrade_dump_globals.sql:32: ERROR:  directory "/tmp/fs/m/demoDataDir-1/16385/1/GPDB_6_301908232"
# already in use as a tablespace
@test "can successful re-run gpupgrade after revert with a tablespace" {
    if ! is_GPDB5 "$GPHOME_SOURCE"; then
      skip "only runs on a GPDB5 source cluster"
    fi

    # create a filespace ....
    local FILESPACE_ROOT=`mktemp -d /tmp/gpupgrade.XXXXXX`
    for dir in master primary1 primary2 primary3 mirror1 mirror2 mirror3; do
      mkdir -p ${FILESPACE_ROOT}/${dir}
    done

    local HOSTNAME=`hostname`
    cat <<EOF > "${FILESPACE_ROOT}/filespace.txt"
filespace:batsFS
${HOSTNAME}:1:${FILESPACE_ROOT}/master/demoDataDir-1
${HOSTNAME}:2:${FILESPACE_ROOT}/primary1/demoDataDir0
${HOSTNAME}:3:${FILESPACE_ROOT}/primary2/demoDataDir1
${HOSTNAME}:4:${FILESPACE_ROOT}/primary3/demoDataDir2
${HOSTNAME}:5:${FILESPACE_ROOT}/mirror1/demoDataDir0
${HOSTNAME}:6:${FILESPACE_ROOT}/mirror2/demoDataDir1
${HOSTNAME}:7:${FILESPACE_ROOT}/mirror3/demoDataDir2
${HOSTNAME}:8:${FILESPACE_ROOT}/master/standby
EOF

    psql -d postgres -c "DROP TABLE IF EXISTS batsTable;"
    psql -d postgres -c "DROP TABLESPACE IF EXISTS batsTbsp;"
    psql -d postgres -c "DROP FILESPACE IF EXISTS batsFS;"

    gpfilespace -c ${FILESPACE_ROOT}/filespace.txt

    psql -d postgres -c "CREATE TABLESPACE batsTbsp FILESPACE batsFS;"
    psql -d postgres -c "CREATE TABLE batsTable(a int) TABLESPACE batsTbsp;"
    psql -d postgres -c "INSERT INTO batsTable SELECT i from generate_series(1,100)i;"

    gpupgrade initialize \
        --source-bindir="${GPHOME_SOURCE}/bin" \
        --target-bindir="${GPHOME_TARGET}/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --mode link \
        --verbose 3>&-

    gpupgrade execute --verbose

    # this is supposed to remove the tablepaces in the target cluster
    # if not, the second execute below will fail
    gpupgrade revert --verbose


    gpupgrade initialize \
        --source-bindir="${GPHOME_SOURCE}/bin" \
        --target-bindir="${GPHOME_TARGET}/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --mode link \
        --verbose 3>&-

    gpupgrade execute --verbose

    # This last revert is used for test cleanup.
    gpupgrade revert --verbose
}