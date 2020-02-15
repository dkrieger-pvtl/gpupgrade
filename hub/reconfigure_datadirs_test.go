package hub_test

import (
	"database/sql"
	"testing"

	"github.com/greenplum-db/gpupgrade/utils"

	"github.com/greenplum-db/gpupgrade/hub"
)

func TestRenameDataDirsInDatabase(t *testing.T) {
	t.Run("it runs against a database", func(t *testing.T) {
		c := hub.MustCreateCluster(t, []utils.SegConfig{
			{
				ContentID: -1,
				Role:      utils.PrimaryRole,
				Port:      6000,
			},
		})

		err := hub.ModifySegmentCatalog(c, []utils.SegConfig{})

		if err != nil {
			t.Fatalf("error: %v", err)
		}
	})

	t.Run("it rolls back on a failure", func(t *testing.T) {
		t.Error("need to handle errors")
	})

	t.Run("it alters a database", func(t *testing.T) {
		c := hub.MustCreateCluster(t, []utils.SegConfig{
			{
				ContentID: -1,
				Role:      utils.PrimaryRole,
				Port:      6000,
			},
		})

		err := hub.ModifySegmentCatalog(c, []utils.SegConfig{
			{DbID: 1, DataDir: "/some/data/dir"},
			{DbID: 2, DataDir: "/some/other/data/dir"},
		})

		if err != nil {
			t.Fatalf("error: %v", err)
		}

		hub.WithinDbTransaction(c, func(transaction *sql.Tx) error {
			dir, err := getDataDir(transaction, 1)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			if dir != "/some/data/dir" {
				t.Errorf("got %v", dir)
			}

			dir, err = getDataDir(transaction, 2)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			if dir != "/some/other/data/dir" {
				t.Errorf("got %v", dir)
			}

			return nil
		})
	})

	t.Run("renames target datadirs by removing _upgrade", func(t *testing.T) {
		// /data/qddir_upgrade/demoDataDir-1 -> datadirs/qddir/demoDataDir-1
		c := hub.MustCreateCluster(t, []utils.SegConfig{
			{
				DbID:    100,
				DataDir: "/data/qddir_upgrade/demoDataDir-1",
				Role:    utils.PrimaryRole,
			},
		})

		dataDirTransformationFunc := func(text string) string {
			return text + "/foo"
		}

		modifications := []utils.SegConfig{}

		spyModifySegmentCatalog := func(_ *utils.Cluster, mods []utils.SegConfig) error {
			modifications = mods
			return nil
		}

		err := hub.ChangeDataDirsInCatalog(c, dataDirTransformationFunc, spyModifySegmentCatalog)

		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		expectedDataDir := "/data/qddir_upgrade/demoDataDir-1/foo"
		if modifications[0].DataDir != expectedDataDir || modifications[0].DbID != 100 {
			t.Errorf("got %q, wanted %v",
				modifications[0].DataDir,
				expectedDataDir)
		}
	})
}

func getDataDir(transaction *sql.Tx, dbid int) (string, error) {
	row := transaction.QueryRow(
		"select datadir from gp_segment_configuration where dbid=$1;",
		dbid)

	value := ""
	err := row.Scan(&value)
	return value, err
}

//func TestRenameDataDirsInDatabase(t *testing.T) {
//c := MustCreateCluster(t, []cluster.SegConfig{
//	{ContentID: -1, DbID: 1, Port: 15432, Hostname: "localhost", DataDir: "/data/qddir/seg-1", Role: "p"},
//	{ContentID: 0, DbID: 2, Port: 25432, Hostname: "host1", DataDir: "/data/dbfast1/seg1", Role: "p"},
//	{ContentID: 1, DbID: 3, Port: 25433, Hostname: "host2", DataDir: "/data/dbfast2/seg2", Role: "p"},
//})
//
//t.Run("5X correctly sets allow_system_table_mods", func(t *testing.T) {
//	c.Version = dbconn.NewVersion("5.24.1")
//
//	var called bool
//	execCommand = exectest.NewCommandWithVerifier(Success, func(path string, args ...string) {
//		called = true
//
//		expected := []string{
//			"--single", "-D /data/qddir/seg-1",
//			"-c \"allow_system_table_mods='DML'\"",
//			"template1",
//		}
//		if !reflect.DeepEqual(args, expected) {
//			t.Errorf("postgres invoked with %q, want %q", args, expected)
//		}
//	})
//
//	err := renameDataDirsInDatabase(c)
//	if err != nil {
//		t.Errorf("unexpected err %#v", err)
//	}
//
//	if !called {
//		t.Errorf("pg_upgrade was not executed")
//	}
//})
//
//t.Run("6X and above correctly sets allow_system_table_mods", func(t *testing.T) {
//	c.Version = dbconn.NewVersion("6.4.0")
//
//	var called bool
//	execCommand = exectest.NewCommandWithVerifier(Success, func(path string, args ...string) {
//		called = true
//
//		expectedArgs := []string{
//			"--single", "-D /data/qddir/seg-1",
//			"-c \"allow_system_table_mods=true\"",
//			"template1",
//		}
//		if !reflect.DeepEqual(args, expectedArgs) {
//			t.Errorf("postgres invoked with %q, want %q", args, expectedArgs)
//		}
//	})
//
//	err := renameDataDirsInDatabase(c)
//	if err != nil {
//		t.Errorf("unexpected err %#v", err)
//	}
//
//	if !called {
//		t.Errorf("pg_upgrade was not executed")
//	}
//})
//
//t.Run("5X uses correct SQL", func(t *testing.T) {
//	c.Version = dbconn.NewVersion("5.24.1")
//
//	//execCommand = exec.Command
//	//defer func() { execCommand = nil }()
//	//
//	//r, _, err := os.Pipe()
//	//if err != nil {
//	//	t.Errorf("unexpected err %#v", err)
//	//}
//	//
//	//cmd := execCommand("postgres")
//	//cmd.Stdin = r
//	//
//	//b, err := ioutil.ReadAll(r)
//	//if err != nil {
//	//	t.Errorf("unexpected err %#v", err)
//	//}
//	//
//	//expected := "toast"
//	//if string(b) != expected {
//	//	t.Errorf("got %q want %q", string(b), expected)
//	//}
//})
//
//t.Run("6X uses correct SQL", func(t *testing.T) {
//	c.Version = dbconn.NewVersion("6.4.0")
//
//})
//
//t.Run("Returns error when postgres fails", func(t *testing.T) {
//	execCommand = exectest.NewCommand(Failure)
//
//	err := renameDataDirsInDatabase(c)
//
//	var multiErr *multierror.Error
//	if !xerrors.As(err, &multiErr) {
//		t.Fatalf("got error %#v, want type %T", err, multiErr)
//	}
//
//	if len(multiErr.Errors) != 1 {
//		t.Errorf("received %d errors, want %d", len(multiErr.Errors), 1)
//	}
//
//	for _, err := range multiErr.Errors {
//		if !strings.Contains(string(err.Error()), "exit status 1") {
//			t.Errorf("wanted error message 'exit status 1' from postgres, got %q", string(err.Error()))
//		}
//	}
//})
//
//t.Run("Returns error when failing to write SQL", func(t *testing.T) {
//	execCommand = exectest.NewCommand(Success)
//
//	expected := errors.New("write failed")
//	utils.System.WriteString = func(w io.Writer, s string) (n int, err error) {
//		return 0, expected
//	}
//
//	err := renameDataDirsInDatabase(c)
//
//	var multiErr *multierror.Error
//	if !xerrors.As(err, &multiErr) {
//		t.Fatalf("got error %#v, want type %T", err, multiErr)
//	}
//
//	if len(multiErr.Errors) != 1 {
//		t.Errorf("received %d errors, want %d", len(multiErr.Errors), 1)
//	}
//
//	for _, err := range multiErr.Errors {
//		if !strings.Contains(string(err.Error()), expected.Error()) {
//			t.Errorf("wanted error message '%q' from psotgres, got %q", expected.Error(), string(err.Error()))
//		}
//	}
//})
//}

//func TestRenameSegmentDataDirsOnDisk(t *testing.T) {
//c := MustCreateCluster(t, []cluster.SegConfig{
//	{ContentID: -1, DbID: 1, Port: 15432, Hostname: "localhost", DataDir: "/data/qddir/seg-1", Role: "p"},
//	{ContentID: 0, DbID: 2, Port: 25432, Hostname: "host1", DataDir: "/data/dbfast1/seg1", Role: "p"},
//	{ContentID: 1, DbID: 3, Port: 25433, Hostname: "host2", DataDir: "/data/dbfast2/seg2", Role: "p"},
//})
//
//t.Run("returns error on failure", func(t *testing.T) {
//	testhelper.SetupTestLogger() // initialize gplog
//
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	client := mock_idl.NewMockAgentClient(ctrl)
//	client.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast1",
//				Dst: "/data/dbfast1",
//			}},
//		},
//	).Return(&idl.ReconfigureDataDirReply{}, nil)
//
//	expected := errors.New("permission denied")
//	failedClient := mock_idl.NewMockAgentClient(ctrl)
//	failedClient.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast2",
//				Dst: "/data/dbfast2",
//			}},
//		},
//	).Return(nil, expected)
//
//	agentConns := []*Connection{
//		{nil, client, "host1", nil},
//		{nil, failedClient, "host2", nil},
//	}
//
//	noop := func(path string) string {
//		return path
//	}
//
//	err := renameSegmentDataDirsOnDisk(agentConns, c, noop, noop)
//
//	var multiErr *multierror.Error
//	if !xerrors.As(err, &multiErr) {
//		t.Fatalf("got error %#v, want type %T", err, multiErr)
//	}
//
//	if len(multiErr.Errors) != 1 {
//		t.Errorf("received %d errors, want %d", len(multiErr.Errors), 1)
//	}
//
//	for _, err := range multiErr.Errors {
//		if !strings.Contains(string(err.Error()), expected.Error()) {
//			t.Errorf("wanted error message '%q' from psotgres, got %q", expected.Error(), string(err.Error()))
//		}
//	}
//})
//
//t.Run("renames target directories", func(t *testing.T) {
//	testhelper.SetupTestLogger() // initialize gplog
//
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	client1 := mock_idl.NewMockAgentClient(ctrl)
//	client1.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast1_upgrade", // XXX: This is correctly failing until we fix our SQL
//				Dst: "/data/dbfast1",
//			}},
//		},
//	).Return(&idl.ReconfigureDataDirReply{}, nil)
//
//	client2 := mock_idl.NewMockAgentClient(ctrl)
//	client2.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast2_upgrade", // XXX: This is correctly failing until we fix our SQL
//				Dst: "/data/dbfast2",
//			}},
//		},
//	).Return(&idl.ReconfigureDataDirReply{}, nil)
//
//	agentConns := []*Connection{
//		{nil, client1, "host1", nil},
//		{nil, client2, "host2", nil},
//	}
//
//	err := renameSegmentDataDirsOnDisk(agentConns, c, upgradeDataDir, noop)
//	if err != nil {
//		t.Errorf("unexpected err %#v", err)
//	}
//})
//
//t.Run("renames source directories", func(t *testing.T) {
//	testhelper.SetupTestLogger() // initialize gplog
//
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	client1 := mock_idl.NewMockAgentClient(ctrl)
//	client1.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast1",
//				Dst: "/data/dbfast1_old", // XXX: This is correctly failing until we fix our SQL
//			}},
//		},
//	).Return(&idl.ReconfigureDataDirReply{}, nil)
//
//	client2 := mock_idl.NewMockAgentClient(ctrl)
//	client2.EXPECT().ReconfigureDataDirectories(
//		gomock.Any(),
//		&idl.ReconfigureDataDirRequest{
//			Pair: []*idl.RenamePair{{
//				Src: "/data/dbfast2",
//				Dst: "/data/dbfast2_old", // XXX: This is correctly failing until we fix our SQL
//			}},
//		},
//	).Return(&idl.ReconfigureDataDirReply{}, nil)
//
//	agentConns := []*Connection{
//		{nil, client1, "host1", nil},
//		{nil, client2, "host2", nil},
//	}
//
//	err := renameSegmentDataDirsOnDisk(agentConns, c, noop, oldDataDir)
//	if err != nil {
//		t.Errorf("unexpected err %#v", err)
//	}
//})
//}
