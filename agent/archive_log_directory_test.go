// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/agent"
	"github.com/greenplum-db/gpupgrade/idl"
)

func TestArchiveLogDirectories(t *testing.T) {
	testhelper.SetupTestLogger()
	server := agent.NewServer(agent.Config{})

	t.Run("bubbles up errors", func(t *testing.T) {
		expected := errors.New("permission denied")

		mockArchiveLogs := func(source, target string) error {
			return expected
		}
		cleanup := agent.SetArchiveLogs(mockArchiveLogs)
		defer cleanup()

		_, err := server.ArchiveLogDirectory(context.Background(), &idl.ArchiveLogDirectoryRequest{})
		if !xerrors.Is(err, expected) {
			t.Errorf("returned error %#v, want %#v", err, expected)
		}
	})

	t.Run("archives log directories", func(t *testing.T) {
		oldLogDir := "/home/gpAdmin/oldlogidr"
		newLogDir := "/home/gpAdmin/newlogdir"
		calls := 0

		mockArchiveLogs := func(source, target string) error {
			calls++

			if source != oldLogDir {
				t.Errorf("got %q want %q", source, oldLogDir)
			}

			if target != newLogDir {
				t.Errorf("got %q want %q", target, newLogDir)
			}

			return nil
		}
		cleanup := agent.SetArchiveLogs(mockArchiveLogs)
		defer cleanup()

		_, err := server.ArchiveLogDirectory(context.Background(), &idl.ArchiveLogDirectoryRequest{OldDir: oldLogDir, NewDir: newLogDir})
		if err != nil {
			t.Errorf("unexpected error %#v", err)
		}

		if calls != 1 {
			t.Errorf("expected rename to be called once, got %d", calls)
		}
	})
}
