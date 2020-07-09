#!/usr/bin/env bats
#
# Copyright (c) 2017-2020 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

load helpers

SCRIPTS_DIR=$BATS_TEST_DIRNAME/../migration_scripts

setup() {
    skip_if_no_gpdb

    PSQL="$GPHOME_SOURCE/bin/psql -X --no-align --tuples-only"

    TEST_DBNAME=testdb
    DEFAULT_DBNAME=postgres
    GPHDFS_USER=gphdfs_user

    $PSQL -c "DROP DATABASE IF EXISTS $TEST_DBNAME;" -d $DEFAULT_DBNAME
    $PSQL -c "DROP ROLE IF EXISTS $GPHDFS_USER;" -d $DEFAULT_DBNAME

    STATE_DIR=`mktemp -d /tmp/gpupgrade.XXXXXX`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"

    gpupgrade kill-services
}

teardown() {
    # XXX Beware, BATS_TEST_SKIPPED is not a documented export.
    if [ -n "${BATS_TEST_SKIPPED}" ]; then
        return
    fi


    if [ -n "$NEW_CLUSTER" ]; then
        delete_finalized_cluster $GPHOME_TARGET $NEW_CLUSTER
    fi

    if [ -n "$MIGRATION_DIR" ]; then
        rm -r $MIGRATION_DIR
    fi

    gpupgrade kill-services
    rm -r "$STATE_DIR"

    restore_cluster

    start_source_cluster

    $GPHOME_SOURCE/bin/psql -c "DROP ROLE IF EXISTS ${GPHDFS_USER}" -d $DEFAULT_DBNAME
    $GPHOME_SOURCE/bin/psql -c "DROP DATABASE IF EXISTS ${TEST_DBNAME}" -d $DEFAULT_DBNAME
}

@test "migration scripts generate sql to modify non-upgradeable objects and fix pg_upgrade check errors" {
    setup_restore_cluster "--mode=copy"

    $PSQL -c "CREATE DATABASE $TEST_DBNAME;" -d $DEFAULT_DBNAME
    $PSQL -f $BATS_TEST_DIRNAME/../migration_scripts/test/create_nonupgradable_objects.sql -d $TEST_DBNAME

    run gpupgrade initialize \
            --source-gphome="$GPHOME_SOURCE" \
            --target-gphome="$GPHOME_TARGET" \
            --source-master-port="${PGPORT}" \
            --temp-port-range 6020-6040 \
            --disk-free-ratio 0 \
            --verbose
    [ "$status" -ne 0 ] || fail "expected initialize to fail due to pg_upgrade check: $output"

    egrep "\"CHECK_UPGRADE\": \"FAILED\"" $GPUPGRADE_HOME/status.json
    egrep "^Checking.*fatal$" $GPUPGRADE_HOME/pg_upgrade/seg-1/pg_upgrade_internal.log

    MIGRATION_DIR=`mktemp -d /tmp/migration.XXXXXX`
    $SCRIPTS_DIR/generate_migration_sql.bash $GPHOME_SOURCE $PGPORT $MIGRATION_DIR

    # the migration script should not remove primary / unique key constraints on partitioned tables, so
    # remove them manually by dropping the table as they can't be dropped.
    $GPHOME_SOURCE/bin/psql -d $TEST_DBNAME -p $PGPORT -f $BATS_TEST_DIRNAME/../migration_scripts/test/drop_not_fixable_objects.sql

    # store the indexes to compare after upgrade
    root_child_indexes_before=$(get_indexes "$GPHOME_SOURCE")

    $SCRIPTS_DIR/execute_migration_sql.bash $GPHOME_SOURCE $PGPORT $MIGRATION_DIR/pre-upgrade

    gpupgrade initialize \
            --source-gphome="$GPHOME_SOURCE" \
            --target-gphome="$GPHOME_TARGET" \
            --source-master-port="${PGPORT}" \
            --temp-port-range 6020-6040 \
            --disk-free-ratio 0 \
            --verbose
    gpupgrade execute --verbose
    gpupgrade finalize --verbose

    $SCRIPTS_DIR/execute_migration_sql.bash $GPHOME_TARGET $PGPORT $MIGRATION_DIR/post-upgrade

    # post-upgrade scripts should create the indexes on the target cluster
    root_child_indexes_after=$(get_indexes "$GPHOME_TARGET")

    # expect the index information to be same after the upgrade
    diff -U3 <(echo "$root_child_indexes_before") <(echo "$root_child_indexes_after")

    NEW_CLUSTER="$MASTER_DATA_DIRECTORY"
}

get_indexes() {
    local gphome=$1
    $gphome/bin/psql -d $TEST_DBNAME -p $PGPORT -Atc "
         SELECT indrelid::regclass, unnest(indkey)
         FROM pg_index pi
         JOIN pg_partition pp ON pi.indrelid=pp.parrelid
         JOIN pg_class pc ON pc.oid=pp.parrelid
         ORDER by 1,2;
        "
    $gphome/bin/psql -d $TEST_DBNAME -p $PGPORT -Atc "
        SELECT indrelid::regclass, unnest(indkey)
        FROM pg_index pi
        JOIN pg_partition_rule pp ON pi.indrelid=pp.parchildrelid
        JOIN pg_class pc ON pc.oid=pp.parchildrelid
        WHERE pc.relhassubclass='f'
        ORDER by 1,2;
    "
}
