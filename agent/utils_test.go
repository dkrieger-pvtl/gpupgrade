package agent_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpupgrade/agent"
	"github.com/greenplum-db/gpupgrade/utils"
)

func TestPostgresOrNonExistent(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("returns nil if dir does not exist", func(t *testing.T) {
		err := agent.PostgresOrNonExistent("/does/not/exist/ABC1234")
		if err != nil {
			t.Errorf("got unexpected err: %v", err)
		}
	})

	t.Run("returns nil if required dirs exist", func(t *testing.T) {

		source, _, tmpDir := setupDataDirs(t)
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		err := agent.PostgresOrNonExistent(source)
		if err != nil {
			t.Errorf("got unexpected err: %v", err)
		}

	})

	t.Run("returns EPERM if dataDir cannot be accessed", func(t *testing.T) {
		expected := &os.PathError{Err: syscall.EPERM}
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, expected
		}
		defer func() {
			utils.System.Stat = os.Stat
		}()

		err := agent.PostgresOrNonExistent("/does/not/matter")

		var multiErr *multierror.Error
		if !xerrors.As(err, &multiErr) {
			t.Fatalf("got error %#v, want type %T", err, multiErr)
		}

		if len(multiErr.Errors) != 1 {
			t.Errorf("got %d errors, want %d", len(multiErr.Errors), 1)
		}

		for _, err := range multiErr.Errors {
			if !xerrors.Is(err, expected) {
				t.Errorf("got error %#v, want %#v", expected, err)
			}
		}
	})

	t.Run("returns EPERM if a postgres file cannot be accessed", func(t *testing.T) {
		source, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer func() {
			os.RemoveAll(source)
		}()

		expected := &os.PathError{Err: syscall.EPERM}
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			if name != filepath.Join(source, "postgresql.conf") {
				return nil, nil
			}
			return nil, expected
		}
		defer func() {
			utils.System.Stat = os.Stat
		}()

		err = agent.PostgresOrNonExistent(source)

		var multiErr *multierror.Error
		if !xerrors.As(err, &multiErr) {
			t.Fatalf("got error %#v, want type %T", err, multiErr)
		}

		if len(multiErr.Errors) != 1 {
			t.Errorf("got %d errors, want %d", len(multiErr.Errors), 1)
		}

		for _, err := range multiErr.Errors {
			if !xerrors.Is(err, expected) {
				t.Errorf("got error %#v, want %#v", expected, err)
			}
		}
	})

}

func setupDataDirs(t *testing.T) (source, target, tmpDir string) {

	var err error
	tmpDir, err = ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	source = createDataDir(t, "source", tmpDir)
	target = createDataDir(t, "target", tmpDir)

	return source, target, tmpDir
}

func createDataDir(t *testing.T, name, tmpDir string) (source string) {

	source = filepath.Join(tmpDir, name)

	err := os.Mkdir(source, 0700)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}

	for _, fileName := range agent.PostgresFiles {
		filePath := filepath.Join(source, fileName)
		err = os.Mkdir(filePath, 0700)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
	}

	return source
}
