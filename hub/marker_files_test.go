package hub_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/greenplum-db/gpupgrade/upgrade"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/hub"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/idl/mock_idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func TestAddMarkerFiles(t *testing.T) {

	upgradeID := upgrade.ID(10) // "CgAAAAAAAAA" in base64

	t.Run("passes correct filenames to OpenFile", func(t *testing.T) {

		conf := new(hub.Config)

		conf.Source = hub.MustCreateCluster(t, []greenplum.SegConfig{
			{ContentID: -1, Hostname: "sdw1", DataDir: "/data/qddir/seg-1", Role: greenplum.PrimaryRole},
			{ContentID: 0, Hostname: "sdw1", DataDir: "/data/dbfast1/seg1", Role: greenplum.PrimaryRole},
			{ContentID: 1, Hostname: "sdw2", DataDir: "/data/dbfast2/seg2", Role: greenplum.PrimaryRole},
			{ContentID: 2, Hostname: "sdw1", DataDir: "/data/dbfast1/seg3", Role: greenplum.PrimaryRole},
			{ContentID: 3, Hostname: "sdw2", DataDir: "/data/dbfast2/seg4", Role: greenplum.PrimaryRole},

			{ContentID: -1, Hostname: "standby", DataDir: "/data/standby", Role: greenplum.MirrorRole},
			{ContentID: 0, Hostname: "sdw1", DataDir: "/data/dbfast_mirror1/seg1", Role: greenplum.MirrorRole},
			{ContentID: 1, Hostname: "sdw2", DataDir: "/data/dbfast_mirror2/seg2", Role: greenplum.MirrorRole},
			{ContentID: 2, Hostname: "sdw1", DataDir: "/data/dbfast_mirror1/seg3", Role: greenplum.MirrorRole},
			{ContentID: 3, Hostname: "sdw2", DataDir: "/data/dbfast_mirror2/seg4", Role: greenplum.MirrorRole},
		})

		conf.Target = hub.MustCreateCluster(t, []greenplum.SegConfig{
			{ContentID: -1, Hostname: "sdw1", DataDir: "/data/qddir/seg-1_CgAAAAAAAAA-1", Role: greenplum.PrimaryRole},
			{ContentID: 0, Hostname: "sdw1", DataDir: "/data/dbfast1/seg1_CgAAAAAAAAA", Role: greenplum.PrimaryRole},
			{ContentID: 1, Hostname: "sdw2", DataDir: "/data/dbfast2/seg2_CgAAAAAAAAA", Role: greenplum.PrimaryRole},
			{ContentID: 2, Hostname: "sdw1", DataDir: "/data/dbfast1/seg3_CgAAAAAAAAA", Role: greenplum.PrimaryRole},
			{ContentID: 3, Hostname: "sdw2", DataDir: "/data/dbfast2/seg4_CgAAAAAAAAA", Role: greenplum.PrimaryRole},

			{ContentID: -1, Hostname: "standby", DataDir: "/data/standby_CgAAAAAAAAA", Role: greenplum.MirrorRole},
			{ContentID: 0, Hostname: "sdw1", DataDir: "/data/dbfast_mirror1/seg1_CgAAAAAAAAA", Role: greenplum.MirrorRole},
			{ContentID: 1, Hostname: "sdw2", DataDir: "/data/dbfast_mirror2/seg2_CgAAAAAAAAA", Role: greenplum.MirrorRole},
			{ContentID: 2, Hostname: "sdw1", DataDir: "/data/dbfast_mirror1/seg3_CgAAAAAAAAA", Role: greenplum.MirrorRole},
			{ContentID: 3, Hostname: "sdw2", DataDir: "/data/dbfast_mirror2/seg4_CgAAAAAAAAA", Role: greenplum.MirrorRole},
		})

		var result []string
		utils.System.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
			result = append(result, name)
			return nil, nil
		}
		defer func() {
			utils.System.OpenFile = os.OpenFile
		}()
		utils.System.Close = func(f *os.File) error {
			return nil
		}
		defer func() {
			utils.System.Close = func(f *os.File) error { return f.Close() }
		}()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// We want the source's primaries and mirrors to be archived, but only
		// the target's upgraded primaries should be moved back to the source
		// locations.
		sdw1 := mock_idl.NewMockAgentClient(ctrl)
		sdw1Strs := []string{
			"/data/dbfast1/seg1/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast1/seg3/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast_mirror1/seg1/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast_mirror1/seg3/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast1/seg1_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast1/seg3_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast_mirror1/seg1_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast_mirror1/seg3_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
		}
		sort.Strings(sdw1Strs)
		expectFiles(sdw1, sdw1Strs)

		sdw2 := mock_idl.NewMockAgentClient(ctrl)
		sdw2Strs := []string{
			"/data/dbfast2/seg2/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast2/seg4/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast_mirror2/seg2/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast_mirror2/seg4/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/dbfast2/seg2_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast2/seg4_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast_mirror2/seg2_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
			"/data/dbfast_mirror2/seg4_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
		}
		sort.Strings(sdw2Strs)
		expectFiles(sdw2, sdw2Strs)

		standby := mock_idl.NewMockAgentClient(ctrl)
		standbyStrs := []string{
			"/data/standby/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/standby_CgAAAAAAAAA/.gpupgrade_TARGET_CgAAAAAAAAA",
		}
		sort.Strings(standbyStrs)
		expectFiles(standby, standbyStrs)

		agentConns := []*hub.Connection{
			{nil, sdw1, "sdw1", nil},
			{nil, sdw2, "sdw2", nil},
			{nil, standby, "standby", nil},
		}

		err := hub.AddMarkerFiles(conf.Source, conf.Target, upgradeID, agentConns)
		if err != nil {
			t.Errorf("AddMarkerFiles() returned error: %+v", err)
		}

		// we only explictly call utils.system.OpenFile for master segments
		// the others go through the mock and hence don't actually make that call.
		var expected []string
		expected = append(expected, []string{
			"/data/qddir/seg-1/.gpupgrade_SOURCE_CgAAAAAAAAA",
			"/data/qddir/seg-1_CgAAAAAAAAA-1/.gpupgrade_TARGET_CgAAAAAAAAA",
		}...)
		sort.Strings(expected)
		sort.Strings(result)
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("expected %v\n got %v", expected, result)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {

		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Errorf("unexpected err: %v", err)
		}
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		tmpPath := func(rel string) string {
			return filepath.Join(tmpDir, rel)
		}

		dataDirSource := tmpPath("qddir/seg-1")
		dataDirTarget := tmpPath("qddir/seg-1_CgAAAAAAAAA-1")
		if err := os.MkdirAll(dataDirSource, 0700); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if err := os.MkdirAll(dataDirTarget, 0700); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		source := hub.MustCreateCluster(t, []greenplum.SegConfig{
			{ContentID: -1, Hostname: "sdw1", DataDir: dataDirSource, Role: greenplum.PrimaryRole},
		})
		target := hub.MustCreateCluster(t, []greenplum.SegConfig{
			{ContentID: -1, Hostname: "sdw1", DataDir: dataDirTarget, Role: greenplum.PrimaryRole},
		})

		// first call should create marker files
		err = hub.AddMarkerFiles(source, target, upgradeID, []*hub.Connection{})
		if err != nil {
			t.Errorf("AddMarkerFiles() returned error: %+v", err)
		}

		_, err = os.Stat(filepath.Join(dataDirSource, ".gpupgrade_SOURCE_CgAAAAAAAAA"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = os.Stat(filepath.Join(dataDirTarget, ".gpupgrade_TARGET_CgAAAAAAAAA"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// second call should not fail and files should still be there
		err = hub.AddMarkerFiles(source, target, upgradeID, []*hub.Connection{})
		if err != nil {
			t.Errorf("AddMarkerFiles() returned error: %+v", err)
		}

		_, err = os.Stat(filepath.Join(dataDirSource, ".gpupgrade_SOURCE_CgAAAAAAAAA"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = os.Stat(filepath.Join(dataDirTarget, ".gpupgrade_TARGET_CgAAAAAAAAA"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

	})

}

// expectFiles is syntactic sugar for setting up an expectation on
// AgentClient.RenameDirectories().
func expectFiles(client *mock_idl.MockAgentClient, files []string) {
	client.EXPECT().CreateFiles(
		gomock.Any(),
		&idl.CreateFilesRequest{Files: files},
	).Return(&idl.CreateFilesReply{}, nil)
}
