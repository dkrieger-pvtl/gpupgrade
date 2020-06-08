package hub_test

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/hub"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/idl/mock_idl"
	"github.com/greenplum-db/gpupgrade/testutils/exectest"
	"github.com/greenplum-db/gpupgrade/utils/rsync"
)

func TestRestoreDataDirectories(t *testing.T) {
	testhelper.SetupTestLogger() // initialize gplog

	t.Run("issues correct rsync commmnds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		rsync.SetRsyncCommand(exectest.NewCommand(hub.Success))
		defer rsync.SetRsyncCommand(exec.Command)

		c := hub.MustCreateCluster(t, []greenplum.SegConfig{
			{ContentID: -1, Hostname: "master", DataDir: "/data/qddir", Role: greenplum.PrimaryRole},
			{ContentID: -1, Hostname: "standby", DataDir: "/data/standby", Role: greenplum.MirrorRole},
			{ContentID: 0, Hostname: "sdw2", DataDir: "/data/dbfast1/seg1", Role: greenplum.PrimaryRole},
			{ContentID: 0, Hostname: "sdw1", DataDir: "/data/dbfast_mirror1/seg1", Role: greenplum.MirrorRole},
			{ContentID: 1, Hostname: "sdw1", DataDir: "/data/dbfast2/seg2", Role: greenplum.PrimaryRole},
			{ContentID: 1, Hostname: "sdw2", DataDir: "/data/dbfast_mirror2/seg2", Role: greenplum.MirrorRole},
		})

		client1 := mock_idl.NewMockAgentClient(ctrl)
		client1.EXPECT().RsyncDataDirectory(
			gomock.Any(),
			&idl.RsyncDataDirectoryRequest{
				Options: []string{"--archive", "--compress", "--stats"},
				Exclude: []string{
					"pg_hba.conf", "postmaster.opts", "postgresql.auto.conf", "internal.auto.conf", "gp_dbid",
					"postgresql.conf", "backup_label.old", "postmaster.pid", "recovery.conf",
				},
				Pairs: []*idl.RsyncDataDirPair{
					{
						Src:         "/data/dbfast_mirror1/seg1" + string(os.PathSeparator),
						DstHostname: "sdw2",
						Dst:         "/data/dbfast1/seg1",
					},
				},
			},
		).Return(&idl.RsyncDataDirectoryReply{}, nil)

		client2 := mock_idl.NewMockAgentClient(ctrl)
		client2.EXPECT().RsyncDataDirectory(
			gomock.Any(),
			&idl.RsyncDataDirectoryRequest{
				Options: []string{"--archive", "--compress", "--stats"},
				Exclude: []string{
					"pg_hba.conf", "postmaster.opts", "postgresql.auto.conf", "internal.auto.conf", "gp_dbid",
					"postgresql.conf", "backup_label.old", "postmaster.pid", "recovery.conf",
				},
				Pairs: []*idl.RsyncDataDirPair{
					{
						Src:         "/data/dbfast_mirror2/seg2" + string(os.PathSeparator),
						DstHostname: "sdw1",
						Dst:         "/data/dbfast2/seg2",
					},
				},
			},
		).Return(&idl.RsyncDataDirectoryReply{}, nil)

		client3 := mock_idl.NewMockAgentClient(ctrl)

		agentConns := []*hub.Connection{
			{nil, client1, "sdw1", nil},
			{nil, client2, "sdw2", nil},
			{nil, client3, "standby", nil},
		}

		err := hub.RestoreMasterAndPrimaries(&noopStream{}, c, agentConns)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}
	})
}

type noopStream struct {
}

func (stream *noopStream) Stdout() io.Writer {
	return ioutil.Discard
}
func (stream *noopStream) Stderr() io.Writer {
	return ioutil.Discard
}
