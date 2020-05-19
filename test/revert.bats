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

    PSQL="$GPHOME"/bin/psql
}

@test "revert cleans up state dir and data dirs" {
    gpupgrade initialize \
        --source-bindir="$GPHOME/bin" \
        --target-bindir="$GPHOME/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    # parse config.json for the datadirs
    local target_hosts_dirs=$(jq -r '.Target.Primaries[] | .Hostname + " " + .DataDir' "${GPUPGRADE_HOME}/config.json")

    # check that the target datadirs exist
    while read -r hostname datadir; do
        ssh "${hostname}" stat "${datadir}" || fail "expected datadir ${datadir} on host ${hostname} to exist"
    done <<< "${target_hosts_dirs}"

    process_is_running "[g]pupgrade hub" || fail 'expected hub to be running'
    process_is_running "[g]pupgrade agent" || fail 'expected agent to be running'

    gpupgrade revert --verbose

    # gpupgrade processes are stopped
    ! process_is_running "[g]pupgrade hub" || fail 'expected hub to have been stopped'
    ! process_is_running "[g]pupgrade agent" || fail 'expected agent to have been stopped'

    # check that the target datadirs were deleted
    while read -r hostname datadir; do
        run ssh "${hostname}" stat "${datadir}"
        ! [ $status -eq 0 ] || fail "expected datadir ${datadir} to have been deleted"
    done <<< "${target_hosts_dirs}"

    # the GPUPGRADE_HOME directory is deleted
    if [ -d "${GPUPGRADE_HOME}" ]; then
        echo "expected GPUPGRADE_HOME directory ${GPUPGRADE_HOME} to have been deleted"
        exit 1
    fi

    # check that archive directory has been created within the last 3 minutes
    NUM_ARCHIVE_DIRS=$(find ${HOME}/gpAdminLogs -type d  -cmin -3 | grep gpupgrade- | wc -l)
    [ "${NUM_ARCHIVE_DIRS}" -gt 0 ] || fail "expected directory matching $HOME/gpAdminLogs/gpupgrade-* to be created"
}

@test "after execute revert stops the target cluster and starts the source cluster" {
    local target_master_port=6020

    gpupgrade initialize \
        --source-bindir="$GPHOME/bin" \
        --target-bindir="$GPHOME/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range ${target_master_port}-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    gpupgrade revert --verbose

    run $PSQL postgres -c "SELECT 1;"
    [ "$status" -eq 0 ] || fail "expected source cluster to be running on port ${PGPORT}"

    run $PSQL postgres -p ${target_master_port} -c "SELECT 1;"
    [ "$status" -ne 0 ] || fail "expected target cluster to not be running on port ${target_master_port}"
}

@test "can successfully run gpupgrade after a revert" {
    gpupgrade initialize \
        --source-bindir="$GPHOME/bin" \
        --target-bindir="$GPHOME/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    gpupgrade revert --verbose

    gpupgrade initialize \
        --source-bindir="$GPHOME/bin" \
        --target-bindir="$GPHOME/bin" \
        --source-master-port="${PGPORT}" \
        --temp-port-range 6020-6040 \
        --disk-free-ratio 0 \
        --verbose 3>&-

    gpupgrade execute --verbose

    # This last revert is used for test cleanup.
    gpupgrade revert --verbose
}
