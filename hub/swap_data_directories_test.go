package hub_test

import (
	"errors"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	"github.com/greenplum-db/gpupgrade/hub"
	"github.com/greenplum-db/gpupgrade/utils"
)

type renameSpy struct {
	calls []*renameCall
}

type renameCall struct {
	originalName string
	newName      string
}

func (s *renameSpy) TimesCalled() int {
	return len(s.calls)
}

func (s *renameSpy) Call(i int) *renameCall {
	return s.calls[i-1]
}

func TestSwapDataDirectories(t *testing.T) {
	testhelper.SetupTestLogger() // init gplog

	afterEach := func() {
		utils.System = utils.InitializeSystemFunctions()
	}

	t.Run("it renames data directories for all source and target data dirs", func(t *testing.T) {
		spy := &renameSpy{}

		utils.System.Rename = spy.renameFunc()

		source := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 99, DataDir: "/some/data/directory", Role: utils.PrimaryRole},
			{ContentID: 100, DataDir: "/some/data/directory/primary1", Role: utils.PrimaryRole},
		})

		target := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 10, DataDir: "/some/qddir_upgrade/dataDirectory", Role: utils.PrimaryRole},
			{ContentID: 100, DataDir: "/some/segment1_upgrade/dataDirectory", Role: utils.PrimaryRole},
		})

		config := &hub.Config{
			Source: source,
			Target: target,
		}

		hub.SwapDataDirectories(config)

		if spy.TimesCalled() != 4 {
			t.Errorf("got Rename called %v times, wanted %v times",
				spy.TimesCalled(),
				4)
		}

		spy.assertDirectoriesMoved(t,
			"/some/data/directory",
			"/some/data/directory_old")

		spy.assertDirectoriesMoved(t,
			"/some/qddir_upgrade/dataDirectory",
			"/some/qddir/dataDirectory")

		spy.assertDirectoriesMoved(t,
			"/some/segment1_upgrade/dataDirectory",
			"/some/segment1/dataDirectory")

		spy.assertDirectoriesMoved(t,
			"/some/data/directory/primary1",
			"/some/data/directory/primary1_old")
	})

	t.Run("it returns an error if the directories cannot be renamed", func(t *testing.T) {
		defer afterEach()

		utils.System.Rename = func(oldpath, newpath string) error {
			return errors.New("failure to rename")
		}

		source := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 99, DataDir: "/some/data/directory", Role: utils.PrimaryRole},
		})

		target := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 99, DataDir: "/some/data/directory", Role: utils.PrimaryRole},
		})

		config := &hub.Config{
			Source: source,
			Target: target,
		}

		err := hub.SwapDataDirectories(config)

		if err == nil {
			t.Fatalf("got nil for an error during SwapDataDirectories, wanted a failure to move directories: %+v", err)
		}
	})

	t.Run("it does not modify the cluster state if there is an error", func(t *testing.T) {
		defer afterEach()

		utils.System.Rename = func(oldpath, newpath string) error {
			return errors.New("failure to rename")
		}

		source := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 99, DataDir: "/some/data/directory", Role: utils.PrimaryRole},
		})

		target := hub.MustCreateCluster(t, []utils.SegConfig{
			{ContentID: 99, DataDir: "/some/data/directory_upgrade", Role: utils.PrimaryRole},
		})

		config := &hub.Config{
			Source: source,
			Target: target,
		}

		err := hub.SwapDataDirectories(config)

		if err == nil {
			t.Fatalf("got nil for an error during SwapDataDirectories, wanted a failure to move directories: %+v", err)
		}

		assertDataDir_NOT_Modified(t,
			config.Source.Primaries[99].DataDir,
			"/some/data/directory",
		)

		assertDataDir_NOT_Modified(t,
			config.Target.Primaries[99].DataDir,
			"/some/data/directory_upgrade",
		)
	})
}

func assertDataDirModified(t *testing.T, newDataDir, expectedDataDir string) {
	if newDataDir != expectedDataDir {
		t.Errorf("got new data dir of %v, wanted %v",
			newDataDir, expectedDataDir)
	}
}

func assertDataDir_NOT_Modified(t *testing.T, newDataDir, expectedDataDir string) {
	if newDataDir != expectedDataDir {
		t.Errorf("got new data dir of %v, wanted %v",
			newDataDir, expectedDataDir)
	}
}

func (spy *renameSpy) assertDirectoriesMoved(t *testing.T, originalName string, newName string) {
	var call *renameCall

	for _, c := range spy.calls {
		if c.originalName == originalName {
			call = c
		}
	}

	if call == nil {
		t.Errorf("got no calls to rename %v, expected 1 call", originalName)
	} else {
		if call.originalName != originalName &&
			call.newName != newName {

			t.Errorf("got rename from %q to %q, wanted rename from %q to %q",
				call.originalName, call.newName,
				originalName, newName)
		}
	}
}

func (spy *renameSpy) renameFunc() func(oldpath string, newpath string) error {
	return func(originalName, newName string) error {
		spy.calls = append(spy.calls, &renameCall{
			originalName: originalName,
			newName:      newName,
		})

		return nil
	}
}
