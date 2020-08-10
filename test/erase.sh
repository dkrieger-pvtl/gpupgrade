#!/bin/bash

source helpers.bash

is_source_standby_in_sync() {
    local START
    local INSYNC="f"

    START=$(date '+%s')

    echo running

    # wait up to 10 minutes for the standby to either disappear or come into sync
    until (( $(date '+%s') - "$START" > 4 )); do
        INSYNC=$("$PSQL" -AXt postgres -c "SELECT sent_location=replay_location from pg_stat_replication")

        if [[ -z "$INSYNC" ]] && ! is_source_standby_running; then
            echo stopped...done
            break
        elif [[ "$INSYNC" == "t" ]]; then
            echo insync...done
            break
        fi
        # otherwise, standby is not in sync yet, but might sync later.

        echo sleeping
        sleep 1
    done

    echo timeout
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

is_source_standby_running


