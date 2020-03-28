package agent_test

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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

	pair := idl.RenamePair{
		Src: "/data/dbfast1_upgrade",
		Dst: "/data/dbfast1",
	}

	req := &idl.RenameDirectoriesRequest{
		Pairs: []*idl.RenamePair{&pair},
	}

	t.Run("successfully renames src and dst data directories", func(t *testing.T) {
		utils.System.Rename = func(src, dst string) error {
			if src != pair.Src {
				t.Errorf("got %q want %q", src, pair.Src)
			}

			if dst != pair.Dst {
				t.Errorf("got %q want %q", dst, pair.Dst)
			}

			return nil
		}
		defer func() {
			utils.System.Rename = os.Rename
		}()

		_, err := server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %#v", err)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
		source := filepath.Join(tmpDir, "source")
		target := filepath.Join(tmpDir, "target")

		err = os.Mkdir(source, 0700)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}

		req := &idl.RenameDirectoriesRequest{
			Pairs: []*idl.RenamePair{
				{
					Src: source,
					Dst: target,
				},
			},
		}

		_, err = server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %#v", err)
		}

		_, err = server.RenameDirectories(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error got %#v", err)
		}

	})

	t.Run("fails to rename src and dst data directories", func(t *testing.T) {
		expected := errors.New("permission denied")
		utils.System.Rename = func(src, dst string) error {
			return expected
		}

		_, err := server.RenameDirectories(context.Background(), req)
		if !xerrors.Is(err, expected) {
			t.Errorf("got %#v want %#v", err, expected)
		}
	})
}
