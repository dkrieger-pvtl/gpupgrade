package services

import (
	"database/sql/driver"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	as "github.com/greenplum-db/gpupgrade/agent/services"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestInitTargetCluster(t *testing.T) {
	g := NewGomegaWithT(t)
	ctrl := gomock.NewController(GinkgoT())
	defer ctrl.Finish()

	hasServerStarted := make(chan bool, 1)
	listener := bufconn.Listen(1024 * 1024)
	agentServer := grpc.NewServer()
	defer agentServer.Stop() // TODO: why does this hang

	idl.RegisterAgentServer(agentServer, &as.AgentServer{})
	go func() {
		hasServerStarted <- true
		if err := agentServer.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	<-hasServerStarted


	_, _, _ = testhelper.SetupTestLogger() // Store gplog output.

	dbConn, sqlMock := testhelper.CreateAndConnectMockDB(1)

	dir, _ := ioutil.TempDir("", "")

	sourceCluster := cluster.NewCluster([]cluster.SegConfig{
		cluster.SegConfig{ContentID: -1, DbID: 1, Port: 15432, Hostname: "localhost", DataDir: fmt.Sprintf("%s/seg-1", dir)},
		cluster.SegConfig{ContentID: 0, DbID: 2, Port: 25432, Hostname: "host1", DataDir: fmt.Sprintf("%s/seg1", dir)},
		cluster.SegConfig{ContentID: 1, DbID: 3, Port: 25433, Hostname: "host2", DataDir: fmt.Sprintf("%s/seg2", dir)},
	})
	source := &utils.Cluster{
		Cluster:    sourceCluster,
		BinDir:     "/source/bindir",
		ConfigPath: "my/config/path",
		Version:    dbconn.GPDBVersion{},
	}

	targetCluster := cluster.NewCluster([]cluster.SegConfig{})
	target := &utils.Cluster{
		Cluster:    targetCluster,
		BinDir:     "/target/bindir",
		ConfigPath: "my/config/path",
		Version:    dbconn.GPDBVersion{},
	}



	hubConf := &HubConfig{
		HubToAgentPort: -1,
		StateDir:       dir,
	}

	hub := NewHub(source, target, hubConf, nil)

	expectedSegDataDirMap := map[string][]string{
		"host1": {fmt.Sprintf("%s_upgrade", dir)},
		"host2": {fmt.Sprintf("%s_upgrade", dir)},
	}

	t.Run("CreateInitialInitsystemConfig: successfully get initial gpinitsystem config array", func(t *testing.T) {
		utils.System.Hostname = func() (string, error) {
			return "mdw", nil
		}
		expectedConfig := []string{
			`ARRAY_NAME="gp_upgrade cluster"`, "SEG_PREFIX=seg",
			"TRUSTED_SHELL=ssh"}
		gpinitsystemConfig, err := hub.CreateInitialInitsystemConfig()
		g.Expect(err).To(BeNil())
		g.Expect(gpinitsystemConfig).To(Equal(expectedConfig))
	})

	t.Run("GetCheckpointSegmentsAndEncoding: successfully get the GUC values", func(t *testing.T) {
		checkpointRow := sqlmock.NewRows([]string{"string"}).AddRow(driver.Value("8"))
		encodingRow := sqlmock.NewRows([]string{"string"}).AddRow(driver.Value("UNICODE"))
		sqlMock.ExpectQuery("SELECT .*checkpoint.*").WillReturnRows(checkpointRow)
		sqlMock.ExpectQuery("SELECT .*server.*").WillReturnRows(encodingRow)
		expectedConfig := []string{"CHECK_POINT_SEGMENTS=8", "ENCODING=UNICODE"}
		gpinitsystemConfig, err := GetCheckpointSegmentsAndEncoding([]string{}, dbConn)
		g.Expect(err).To(BeNil())
		g.Expect(gpinitsystemConfig).To(Equal(expectedConfig))
	})

	t.Run("DeclareDataDirectories: successfully declares all directories", func(t *testing.T) {
		expectedConfig := []string{fmt.Sprintf("QD_PRIMARY_ARRAY=localhost~15433~%[1]s_upgrade/seg-1~1~-1~0", dir),
			fmt.Sprintf(`declare -a PRIMARY_ARRAY=(
	host1~29432~%[1]s_upgrade/seg1~2~0~0
	host2~29433~%[1]s_upgrade/seg2~3~1~0
)`, dir)}
		resultConfig, resultMap, port := hub.DeclareDataDirectories([]string{})
		g.Expect(resultMap).To(Equal(expectedSegDataDirMap))
		g.Expect(resultConfig).To(Equal(expectedConfig))
		g.Expect(port).To(Equal(15433))
	})

	t.Run("CreateAllDataDirectories: successfully creates all directories", func(t *testing.T) {
		statCalls := []string{}
		mkdirCalls := []string{}
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			statCalls = append(statCalls, name)
			return nil, os.ErrNotExist
		}
		utils.System.MkdirAll = func(path string, perm os.FileMode) error {
			mkdirCalls = append(mkdirCalls, path)
			return nil
		}
		fakeConns := []*Connection{}
		err := hub.CreateAllDataDirectories(fakeConns, expectedSegDataDirMap)
		g.Expect(err).To(BeNil())
		g.Expect(statCalls).To(Equal([]string{fmt.Sprintf("%s_upgrade", dir)}))
		g.Expect(mkdirCalls).To(Equal([]string{fmt.Sprintf("%s_upgrade", dir)}))
	})

	t.Run("CreateAllDataDirectories: cannot stat the master data directory", func(t *testing.T) {
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, errors.New("permission denied")
		}
		fakeConns := []*Connection{}
		expectedErr := errors.Errorf("Error statting new directory %s_upgrade: permission denied", dir)
		err := hub.CreateAllDataDirectories(fakeConns, expectedSegDataDirMap)
		g.Expect(err.Error()).To(Equal(expectedErr.Error()))
	})

	t.Run("CreateAllDataDirectories: cannot create the master data directory", func(t *testing.T) {
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		utils.System.MkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("permission denied")
		}
		fakeConns := []*Connection{}
		expectedErr := errors.New("Could not create new directory: permission denied")
		err := hub.CreateAllDataDirectories(fakeConns, expectedSegDataDirMap)
		g.Expect(err.Error()).To(Equal(expectedErr.Error()))
	})

	t.Run("CreateAllDataDirectories: cannot create the segment data directories", func(t *testing.T) {
		badConnection, err := grpc.Dial("nonExistHost", grpc.WithInsecure())
		g.Expect(err).To(Not(HaveOccurred()))

		agentConns := []*Connection{{"nonExistHost:123", badConnection, func() {}}}

		err = hub.CreateAllDataDirectories(agentConns, expectedSegDataDirMap)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("Error creating segment data directories"))
	})

	t.Run("RunInitsystemForTargetCluster", func(t *testing.T) {
		// XXX: See other test file that is in package services which enables us to test execCommand.
	})

	t.Run("GetMasterSegPrefix", func(t *testing.T) {
		DescribeTable("returns a valid seg prefix given",
			func(datadir string) {
				segPrefix, err := GetMasterSegPrefix(datadir)
				g.Expect(segPrefix).To(Equal("gpseg"))
				g.Expect(err).ShouldNot(HaveOccurred())
			},
			Entry("an absolute path", "/data/master/gpseg-1"),
			Entry("a relative path", "../master/gpseg-1"),
			Entry("a implicitly relative path", "gpseg-1"),
		)

		DescribeTable("returns errors when given",
			func(datadir string) {
				_, err := GetMasterSegPrefix(datadir)
				g.Expect(err).To(HaveOccurred())
			},
			Entry("the empty string", ""),
			Entry("a path without a content identifier", "/opt/myseg"),
			Entry("a path with a segment content identifier", "/opt/myseg2"),
			Entry("a path that is only a content identifier", "-1"),
			Entry("a path that ends in only a content identifier", "///-1"),
		)
	})
}
