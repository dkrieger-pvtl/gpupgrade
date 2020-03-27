package agent_test

import (
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/agent"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func TestRenameDirectories(t *testing.T) {
	testhelper.SetupTestLogger()

	server := agent.NewServer(agent.Config{
		Port:     -1,
		StateDir: "",
	})

	t.Run("calls rename with correct src and dst data directories", func(t *testing.T) {

		source, target, tmpDir := setupDataDirs(t)
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		pair := idl.RenamePair{
			Src: source,
			Dst: target,
		}

		req := &idl.RenameDirectoriesRequest{
			Pairs: []*idl.RenamePair{
				&pair,
			},
		}

		called := false
		utils.System.Rename = func(src, dst string) error {
			if src != pair.Src {
				t.Errorf("got %q want %q", src, pair.Src)
			}

			if dst != pair.Dst {
				t.Errorf("got %q want %q", dst, pair.Dst)
			}

			called = true
			return nil
		}
		defer func() {
			utils.System.Rename = os.Rename
		}()

		_, err := server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %#v", err)
		}
		if !called {
			t.Errorf("unexpected true, got false")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {

		source, target, tmpDir := setupDataDirs(t)
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		req := &idl.RenameDirectoriesRequest{
			Pairs: []*idl.RenamePair{
				{
					Src: source,
					Dst: target,
				},
			},
		}

		_, err := server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %v", err)
		}

		_, err = server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %v", err)
		}

	})

	t.Run("fails when rename fails with a EPERM error", func(t *testing.T) {

		source, target, tmpDir := setupDataDirs(t)
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		req := &idl.RenameDirectoriesRequest{
			Pairs: []*idl.RenamePair{
				{
					Src: source,
					Dst: target,
				},
			},
		}

		expected := &os.LinkError{Err: syscall.EPERM}
		utils.System.Rename = func(src, dst string) error {
			return expected
		}

		_, err := server.RenameDirectories(context.Background(), req)
		if !xerrors.Is(err, expected) {
			t.Errorf("got %#v want %#v", err, expected)
		}
	})
}
