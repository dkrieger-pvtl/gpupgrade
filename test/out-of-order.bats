#! /usr/bin/env bats
#
# This file provides negative test cases for when the user does not execute
# upgrade steps in the correct order after starting the hub.

load helpers

setup() {
    STATE_DIR=`mktemp -d`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"

    gpupgrade initialize --old-bindir /usr/local/gpdb6/bin/ --new-bindir /usr/local/gpdb6/bin/
}

teardown() {
    gpupgrade stop
    rm -r "${STATE_DIR}"
}

# todo: add tests
