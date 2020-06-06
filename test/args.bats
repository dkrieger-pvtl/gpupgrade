#! /usr/bin/env bats
#
# Copyright (c) 2017-2020 VMware, Inc. or its affiliates
# SPDX-License-Identifier: Apache-2.0

load helpers

setup() {
    skip_if_no_gpdb

    STATE_DIR=`mktemp -d /tmp/gpupgrade.XXXXXX`
    export GPUPGRADE_HOME="${STATE_DIR}/gpupgrade"

    gpupgrade kill-services
}

teardown() {
    gpupgrade kill-services
    rm -r "$STATE_DIR"
}

@test "gpupgrade subcommands fail when passed insufficient arguments" {
    run gpupgrade initialize
    [ "$status" -eq 1 ]
    if ! [[ "$output" = *'Required flag(s) "source-bindir", "source-master-port", "target-bindir" have/has not been set'* ]]; then
        fail "actual: $output"
    fi

    run gpupgrade config set
    [ "$status" -eq 1 ]
    if ! [[ "$output" = *'the set command requires at least one flag to be specified'* ]]; then
        fail "actual: $output"
    fi
}

@test "gpupgrade initialize fails when other flags are used with --file" {
    run gpupgrade initialize --file /some/config --source-bindir /old/bindir
    [ "$status" -eq 1 ]
    if ! [[ "$output" = *'--file cannot be used with any other flag'* ]]; then
        fail "actual: $output"
    fi
}

@test "gpupgrade initialize --file uses configured values" {
    config_file=${STATE_DIR}/gpupgrade_config
    cat <<- EOF > "$config_file"
		source-bindir = /my/old/bin/dir
		target-bindir = /my/new/bin/dir
		source-master-port = ${PGPORT}
		disk-free-ratio = 0
		stop-before-cluster-creation = true
	EOF

    gpupgrade initialize --file "$config_file"

    run gpupgrade config show
    [ "$status" -eq 0 ]
    [[ "${lines[0]}" = "id - "* ]] # this is randomly generated; we could replace * with a base64 regex matcher
    [ "${lines[1]}" = "source-bindir - /my/old/bin/dir" ]
    [ "${lines[2]}" = "target-bindir - /my/new/bin/dir" ]
    [ "${lines[3]}" = "target-datadir - " ] # This isn't populated until cluster creation, but it's still displayed here
}
