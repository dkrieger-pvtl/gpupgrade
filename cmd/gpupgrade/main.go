// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	_ "github.com/lib/pq"

	"github.com/greenplum-db/gpupgrade/cli/commands"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/greenplum-db/gpupgrade/utils/daemon"
)

func main() {
	debug.SetTraceback("all")
	logdir, err := utils.GetLogDir()
	if err != nil {
		fmt.Printf("\n%+v\n", err)
		os.Exit(1)
	}
	gplog.InitializeLogging("gpupgrade_cli", logdir)

	root := commands.BuildRootCommand()
	// Silence usage since Cobra prints usage for all errors rather than just
	// "unknown flag" errors.
	root.SilenceUsage = true

	err = root.Execute()
	if err != nil && err != daemon.ErrSuccessfullyDaemonized {
		if strings.HasPrefix(err.Error(), "unknown flag") {
			cmd := os.Args[1]
			fmt.Println(commands.Help[cmd])
		}

		// We use gplog.Debug instead of Error so the error is not displayed
		// twice to the user in the terminal.
		gplog.Debug("%+v", err)
		os.Exit(1)
	}
}
