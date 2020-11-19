// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub

import (
	"fmt"
	"strings"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/greenplum-db/gpupgrade/utils/errorlist"
)

var GetGpupgradeVersionFunc = GetGpupgradeVersion

type HostVersionInfo struct {
	host             string
	gpupgradeVersion string
	err              error
}

func VerifyGpupgradeAndGPDBVersionsAcrossHosts(agentHosts []string, hubHost string) error {
	hubGpupgradeVersion, err := GetGpupgradeVersionFunc(hubHost)
	if err != nil {
		return xerrors.Errorf("getting hub version: %w", err)
	}

	var wg sync.WaitGroup
	versionChan := make(chan HostVersionInfo, len(agentHosts))

	for _, host := range agentHosts {
		wg.Add(1)

		go func(host string) {
			defer wg.Done()

			gpupgradeVersion, err := GetGpupgradeVersionFunc(host)
			versionChan <- HostVersionInfo{host: host, gpupgradeVersion: gpupgradeVersion, err: err}
		}(host)
	}

	wg.Wait()
	close(versionChan)

	var errs error
	gpupgradeVersionToHosts := make(map[string][]string)
	for agent := range versionChan {
		errs = errorlist.Append(errs, agent.err)

		if hubGpupgradeVersion != agent.gpupgradeVersion {
			gpupgradeVersionToHosts[agent.gpupgradeVersion] = append(gpupgradeVersionToHosts[agent.gpupgradeVersion], agent.host)
		}
	}

	if errs != nil {
		return errs
	}

	if len(gpupgradeVersionToHosts) != 0 {
		var text string
		for gpupgradeVersion, hosts := range gpupgradeVersionToHosts {
			text += fmt.Sprintf("%q: %s\n", gpupgradeVersion, strings.Join(hosts, ", "))
		}

		return xerrors.Errorf(`Version mismatch between gpupgrade hub and agent hosts. 
Hub version: %q

Mismatched Agents:
%s`, hubGpupgradeVersion, text)
	}

	return nil
}

func GetGpupgradeVersion(host string) (string, error) {
	gpupgradePath, err := utils.GetGpupgradePath()
	if err != nil {
		return "", xerrors.Errorf("getting gpupgrade binary path: %w", err)
	}

	cmd := execCommand("ssh", host, fmt.Sprintf(`bash -c "%s version"`, gpupgradePath))
	gplog.Debug("running cmd %q", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", xerrors.Errorf("%q failed with %q: %w", cmd.String(), string(output), err)
	}

	gplog.Debug("output: %q", output)

	return string(output), nil
}
