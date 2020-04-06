package agent_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpupgrade/agent"
	"github.com/greenplum-db/gpupgrade/idl"
)

func TestCreateFiles(t *testing.T) {
	testhelper.SetupTestLogger()

	server := agent.NewServer(agent.Config{
		Port:     -1,
		StateDir: "",
	})

	t.Run("calls OpenFile on correct files", func(t *testing.T) {

		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		expected := []string{filepath.Join(tmpDir, "dataDir1"), filepath.Join(tmpDir, "dataDir2")}
		req := &idl.CreateFilesRequest{Files: expected}

		_, err = server.CreateFiles(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}

		// on first pass, files should be created.
		for _, expect := range expected {
			_, err := os.Stat(expect)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}

		//	on second pass, should also work
		_, err = server.CreateFiles(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}

	})

}
