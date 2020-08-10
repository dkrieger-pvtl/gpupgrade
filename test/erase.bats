#!/usr/bin/env bats

load helpers


# gp_segment_configuration does not show us the status correctly. We must check that the
# sent_location from the master equals the replay_location of the standby.
is_source_standby_in_sync() {
    local INSYNC="f"
    local duration=4 # wait up to 10 minutes
    local poll=1

    echo "running..." >&3

    while (( duration > 0 )); do
        INSYNC=$("$PSQL" -AXt postgres -c "SELECT sent_location=replay_location FROM pg_stat_replication")

        if [[ -z "$INSYNC" ]] && ! is_source_standby_running; then
            echo standbyGone >&3
            break # standby has disappeared
        elif [[ "$INSYNC" == "t" ]]; then
             echo insync >&3
            break
        fi

        echo sleeping... >&3
        sleep $poll
        (( duration = duration - poll ))
    done

    [[ $INSYNC == "t" ]]
}

# use "pg_ctl -D <standby_data_dir> " to determine if the standby is running
is_source_standby_running() {
    local standby_datadir
    standby_datadir=$(query_datadirs "$GPHOME_SOURCE" "$PGPORT" "content = '-1' AND role = 'm'")

    if ! "${GPHOME_SOURCE}"/bin/pg_ctl status -D "$standby_datadir" > /dev/null; then
        return 1
    fi
}

GPHOME_SOURCE=/usr/local/gpdb5
PSQL=/usr/local/gpdb5/bin/psql


@test "experiment" {
    is_source_standby_in_sync
}