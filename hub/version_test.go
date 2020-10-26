// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub_test

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/greenplum-db/gpupgrade/hub"
	"github.com/greenplum-db/gpupgrade/testutils/exectest"
	"github.com/greenplum-db/gpupgrade/testutils/testlog"
)

func gpupgradeVersion() {}

func init() {
	exectest.RegisterMains(
		gpupgradeVersion,
	)
}

func ResetGetVersion() {
	hub.GetVersionFunc = hub.GetVersion
}

func TestValidateGpupgradeVersion(t *testing.T) {
	testlog.SetupLogger()

	hub.SetExecCommand(exectest.NewCommand(gpupgradeVersion))
	defer hub.ResetExecCommand()

	agentHosts := []string{"sdw1", "sdw2"}
	hubHost := "mdw"

	t.Run("ValidateGpupgradeVersion successfully validates the version of gpupgrade on hub and agents", func(t *testing.T) {
		var expectedArgs []string
		for _, host := range append(agentHosts, hubHost) {
			expectedArgs = append(expectedArgs, fmt.Sprintf(`%s bash -c "%s/gpupgrade version"`, host, mustGetExecutablePath(t)))
		}

		var actualArgs []string
		execCmd := exectest.NewCommandWithVerifier(gpupgradeVersion, func(name string, args ...string) {
			if name != "ssh" {
				t.Errorf("execCommand got %q want ssh", name)
			}

			actualArgs = append(actualArgs, strings.Join(args, " "))
		})

		hub.SetExecCommand(execCmd)
		defer hub.ResetExecCommand()

		err := hub.ValidateGpupgradeVersion(hubHost, agentHosts)
		if err != nil {
			t.Errorf("unexpected errr %#v", err)
		}

		sort.Strings(actualArgs)
		sort.Strings(expectedArgs)
		if !reflect.DeepEqual(actualArgs, expectedArgs) {
			t.Errorf("got %q, want %q", actualArgs, expectedArgs)
		}
	})

	t.Run("errors when execCommand fails", func(t *testing.T) {
		hub.SetExecCommand(exectest.NewCommand(hub.Failure))
		defer hub.ResetExecCommand()

		err := hub.ValidateGpupgradeVersion(hubHost, agentHosts)
		if err == nil {
			t.Errorf("expected an error")
		}
	})

	t.Run("reports version mismatch between hub and agent", func(t *testing.T) {
		hub.GetVersionFunc = func(host, path string) (string, error) {
			if host == hubHost {
				return "Version: 0.4.0 Commit: 21b66d7 Release: Dev Build", nil
			}
			return "Version: 0.3.0 Commit: 22b77d7 Release: Dev Build", nil
		}
		defer ResetGetVersion()

		err := hub.ValidateGpupgradeVersion(hubHost, agentHosts)
		if err == nil {
			t.Errorf("expected an error")
		}

		expectedRegex := regexp.MustCompile(`Mismatched Agents:.*\nsdw1.*\nsdw2`)
		if !expectedRegex.Match([]byte(err.Error())) {
			t.Errorf("expected sdw1 and sdw2 in mismatched agents, got %s", err)
		}
	})
}
