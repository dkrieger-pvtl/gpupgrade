// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

var GetVersionFunc = GetVersion

func ValidateGpupgradeVersion(hubHost string, agentHosts []string) error {
	path, err := getBinaryPath()
	if err != nil {
		return xerrors.Errorf("getting gpupgrade binary path: %w", err)
	}

	hubVersion, err := GetVersionFunc(hubHost, path)
	if err != nil {
		return xerrors.Errorf("getting hub version: %w", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(agentHosts))
	mismatchedHostsChan := make(chan string, len(agentHosts))

	for _, host := range agentHosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			version, err := GetVersionFunc(host, path)
			if err != nil {
				errs <- err
				return
			}

			if hubVersion != version {
				mismatchedHostsChan <- host
			}
		}(host)
	}

	wg.Wait()
	close(errs)
	close(mismatchedHostsChan)

	var mismatchedHosts []string
	for host := range mismatchedHostsChan {
		mismatchedHosts = append(mismatchedHosts, host)
	}

	if len(mismatchedHosts) == 0 {
		return nil
	}

	sort.Strings(mismatchedHosts)
	return xerrors.Errorf(`Version mismatch between gpupgrade hub and agent hosts. Found hub version:
%s

Agents with mismatched version: %s`, hubVersion, strings.Join(mismatchedHosts, ", "))
}

func GetVersion(host, path string) (string, error) {
	cmd := execCommand("ssh", host, fmt.Sprintf(`bash -c "%s version"`, path))
	gplog.Debug("running cmd %q", cmd.String())
	version, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if xerrors.As(err, &exitErr) {
			return "", xerrors.Errorf("%q failed with %q: %w", cmd.String(), exitErr.Stderr, err)
		}
		return "", xerrors.Errorf("%q failed with: %w", cmd.String(), err)
	}

	gplog.Debug("output: %q", version)

	return string(version), nil
}
