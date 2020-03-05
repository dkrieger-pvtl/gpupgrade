package utils_test

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"

	"github.com/greenplum-db/gpupgrade/testutils"
	"github.com/greenplum-db/gpupgrade/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"golang.org/x/xerrors"
)

func TestCluster(t *testing.T) {
	primaries := map[int]utils.SegConfig{
		-1: {DbID: 1, ContentID: -1, Port: 5432, Hostname: "localhost", DataDir: "/data/gpseg-1"},
		0:  {DbID: 2, ContentID: 0, Port: 20000, Hostname: "localhost", DataDir: "/data/gpseg0"},
		2:  {DbID: 4, ContentID: 2, Port: 20002, Hostname: "localhost", DataDir: "/data/gpseg2"},
		3:  {DbID: 5, ContentID: 3, Port: 20003, Hostname: "remotehost2", DataDir: "/data/gpseg3"},
	}
	for content, seg := range primaries {
		seg.Role = utils.PrimaryRole
		seg.PreferredRole = utils.PrimaryRole
		primaries[content] = seg
	}

	mirrors := map[int]utils.SegConfig{
		-1: {DbID: 8, ContentID: -1, Port: 5433, Hostname: "localhost", DataDir: "/mirror/gpseg-1"},
		0:  {DbID: 3, ContentID: 0, Port: 20001, Hostname: "localhost", DataDir: "/mirror/gpseg0"},
		2:  {DbID: 6, ContentID: 2, Port: 20004, Hostname: "localhost", DataDir: "/mirror/gpseg2"},
		3:  {DbID: 7, ContentID: 3, Port: 20005, Hostname: "remotehost2", DataDir: "/mirror/gpseg3"},
	}
	for content, seg := range mirrors {
		seg.Role = utils.MirrorRole
		seg.PreferredRole = utils.MirrorRole
		mirrors[content] = seg
	}

	master := primaries[-1]
	standby := mirrors[-1]

	cases := []struct {
		name      string
		primaries []utils.SegConfig
		mirrors   []utils.SegConfig
	}{
		{"mirrorless single-host, single-segment", []utils.SegConfig{master, primaries[0]}, nil},
		{"mirrorless single-host, multi-segment", []utils.SegConfig{master, primaries[0], primaries[2]}, nil},
		{"mirrorless multi-host, multi-segment", []utils.SegConfig{master, primaries[0], primaries[3]}, nil},
		{"single-host, single-segment",
			[]utils.SegConfig{master, primaries[0]},
			[]utils.SegConfig{mirrors[0]},
		},
		{"single-host, multi-segment",
			[]utils.SegConfig{master, primaries[0], primaries[2]},
			[]utils.SegConfig{mirrors[0], mirrors[2]},
		},
		{"multi-host, multi-segment",
			[]utils.SegConfig{master, primaries[0], primaries[3]},
			[]utils.SegConfig{mirrors[0], mirrors[3]},
		},
		{"multi-host, multi-segment with standby",
			[]utils.SegConfig{master, primaries[0], primaries[3]},
			[]utils.SegConfig{standby, mirrors[0], mirrors[3]},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s cluster", c.name), func(t *testing.T) {
			segments := append(c.primaries, c.mirrors...)

			actualCluster, err := utils.NewCluster(segments)
			if err != nil {
				t.Fatalf("returned error %+v", err)
			}

			actualContents := actualCluster.GetContentList()

			var expectedContents []int
			for _, p := range c.primaries {
				expectedContents = append(expectedContents, p.ContentID)
			}

			if !reflect.DeepEqual(actualContents, expectedContents) {
				t.Errorf("GetContentList() = %v, want %v", actualContents, expectedContents)
			}

			for _, expected := range c.primaries {
				content := expected.ContentID

				actual := actualCluster.Primaries[content]
				if actual != expected {
					t.Errorf("Primaries[%d] = %+v, want %+v", content, actual, expected)
				}

				host := actualCluster.GetHostForContent(content)
				if host != expected.Hostname {
					t.Errorf("GetHostForContent(%d) = %q, want %q", content, host, expected.Hostname)
				}

				port := actualCluster.GetPortForContent(content)
				if port != expected.Port {
					t.Errorf("GetPortForContent(%d) = %d, want %d", content, port, expected.Port)
				}

				dbid := actualCluster.GetDbidForContent(content)
				if dbid != expected.DbID {
					t.Errorf("GetDbidForContent(%d) = %d, want %d", content, dbid, expected.DbID)
				}

				datadir := actualCluster.GetDirForContent(content)
				if datadir != expected.DataDir {
					t.Errorf("GetDirForContent(%d) = %q, want %q", content, datadir, expected.DataDir)
				}
			}

			for _, expected := range c.mirrors {
				content := expected.ContentID

				actual := actualCluster.Mirrors[content]
				if actual != expected {
					t.Errorf("Mirrors[%d] = %+v, want %+v", content, actual, expected)
				}
			}
		})
	}

	errCases := []struct {
		name     string
		segments []utils.SegConfig
	}{
		{"bad role", []utils.SegConfig{
			{Role: "x"},
		}},
		{"mirror switched role to primary", []utils.SegConfig{
			{Role: "p", PreferredRole: "m"},
		}},
		{"primary switched role to mirror", []utils.SegConfig{
			{Role: "m", PreferredRole: "p"},
		}},
		{"mirror without primary", []utils.SegConfig{
			{ContentID: 0, Role: "p", PreferredRole: "p"},
			{ContentID: 1, Role: "m", PreferredRole: "m"},
		}},
		{"duplicated primary contents", []utils.SegConfig{
			{ContentID: 0, Role: "p", PreferredRole: "p"},
			{ContentID: 0, Role: "p", PreferredRole: "p"},
		}},
		{"duplicated mirror contents", []utils.SegConfig{
			{ContentID: 0, Role: "p", PreferredRole: "p"},
			{ContentID: 0, Role: "m", PreferredRole: "m"},
			{ContentID: 0, Role: "m", PreferredRole: "m"},
		}},
	}

	for _, c := range errCases {
		t.Run(fmt.Sprintf("doesn't allow %s", c.name), func(t *testing.T) {
			_, err := utils.NewCluster(c.segments)

			if !xerrors.Is(err, utils.ErrInvalidSegments) {
				t.Errorf("returned error %#v, want %#v", err, utils.ErrInvalidSegments)
			}
		})
	}
}

func TestGetSegmentConfiguration(t *testing.T) {
	testhelper.SetupTestLogger() // init gplog

	cases := []struct {
		name     string
		rows     [][]driver.Value
		expected []utils.SegConfig
	}{{
		"single-host, single-segment",
		[][]driver.Value{
			{"0", "localhost", "/data/gpseg0"},
		},
		[]utils.SegConfig{
			{ContentID: 0, Hostname: "localhost", DataDir: "/data/gpseg0"},
		},
	}, {
		"single-host, multi-segment",
		[][]driver.Value{
			{"0", "localhost", "/data/gpseg0"},
			{"1", "localhost", "/data/gpseg1"},
		},
		[]utils.SegConfig{
			{ContentID: 0, Hostname: "localhost", DataDir: "/data/gpseg0"},
			{ContentID: 1, Hostname: "localhost", DataDir: "/data/gpseg1"},
		},
	}, {
		"multi-host, multi-segment",
		[][]driver.Value{
			{"0", "localhost", "/data/gpseg0"},
			{"1", "localhost", "/data/gpseg1"},
			{"2", "remotehost", "/data/gpseg2"},
		},
		[]utils.SegConfig{
			{ContentID: 0, Hostname: "localhost", DataDir: "/data/gpseg0"},
			{ContentID: 1, Hostname: "localhost", DataDir: "/data/gpseg1"},
			{ContentID: 2, Hostname: "remotehost", DataDir: "/data/gpseg2"},
		},
	}}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s cluster", c.name), func(t *testing.T) {
			// Set up the connection to return the expected rows.
			rows := sqlmock.NewRows([]string{"contentid", "hostname", "datadir"})
			for _, row := range c.rows {
				rows.AddRow(row...)
			}

			connection, mock := testhelper.CreateAndConnectMockDB(1)
			mock.ExpectQuery("SELECT (.*)").WillReturnRows(rows)
			defer func() {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Errorf("%v", err)
				}
			}()

			results, err := utils.GetSegmentConfiguration(connection)
			if err != nil {
				t.Errorf("returned error %+v", err)
			}

			if !reflect.DeepEqual(results, c.expected) {
				t.Errorf("got configuration %+v, want %+v", results, c.expected)
			}
		})
	}
}

func TestGetHostnamesExcludingMaster(t *testing.T) {
	testStateDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("got error when creating tempdir: %+v", err)
	}
	expectedCluster := testutils.CreateMultinodeSampleCluster("/tmp")
	expectedCluster.Mirrors = map[int]utils.SegConfig{
		-1: {ContentID: -1, DbID: 1, Port: 15433, Hostname: "standby-host", DataDir: "/seg-1"},
		0:  {ContentID: 0, DbID: 2, Port: 25434, Hostname: "mirror-host1", DataDir: "/seg1"},
		1:  {ContentID: 1, DbID: 3, Port: 25435, Hostname: "mirror-host2", DataDir: "/seg2"},
	}
	expectedCluster.BinDir = "/fake/path"
	expectedCluster.Version = dbconn.NewVersion("6.0.0")
	testhelper.SetupTestLogger()

	defer func() {
		os.RemoveAll(testStateDir)
	}()

	t.Run("returns a list of hosts for only the primaries", func(t *testing.T) {
		actual := expectedCluster.GetHostnamesExcludingMaster(false)
		sort.Strings(actual)

		expected := []string{"host1", "host2"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected hostnames: %#v got: %#v", expected, actual)
		}
	})

	t.Run("returns a list of hosts including mirrors", func(t *testing.T) {
		actual := expectedCluster.GetHostnamesExcludingMaster(true)
		sort.Strings(actual)

		expected := []string{"host1", "host2", "standby-host"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected hostnames: %#v got: %#v", expected, actual)
		}
	})
}

func TestSegmentsOnExcludingMaster(t *testing.T) {
	t.Run("errors when host is not found", func(t *testing.T) {
		cluster := utils.Cluster{}
		_, err := cluster.SegmentsOnExcludingMaster("unknownHost", false)

		expected := utils.UnknownHostError{Hostname: "unknownHost"}
		if !xerrors.Is(err, expected) {
			t.Errorf("returned error %#v, want %#v", err, expected)
		}
	})

	cluster := &utils.Cluster{
		ContentIDs: []int{-1, 0, 1, 2, 3},
		Primaries: map[int]utils.SegConfig{
			-1: {ContentID: -1, Hostname: "host", Role: "p"},
			0:  {ContentID: 0, Hostname: "host", Role: "p"},
			1:  {ContentID: 1, Hostname: "host", Role: "p"},
			2:  {ContentID: 2, Hostname: "excludedHost", Role: "p"},
			3:  {ContentID: 3, Hostname: "excludedHost", Role: "p"},
		},
		Mirrors: map[int]utils.SegConfig{
			-1: {ContentID: -1, Hostname: "host", Role: "m"},
			0:  {ContentID: 0, Hostname: "host", Role: "m"},
			1:  {ContentID: 1, Hostname: "host", Role: "m"},
			2:  {ContentID: 2, Hostname: "excludedHost", Role: "m"},
			3:  {ContentID: 3, Hostname: "excludedHost", Role: "m"},
		},
	}

	cases := []struct {
		name           string
		includeMirrors bool
		expected       []utils.SegConfig
	}{
		{"returns correct segments when master is on same host as a primary and mirror and includeMirrors is false",
			false,
			[]utils.SegConfig{
				{ContentID: 0, Hostname: "host", Role: "p"},
				{ContentID: 1, Hostname: "host", Role: "p"},
			},
		},
		{"returns correct segments master is on same host as a primary and a mirror and includeMirrors is true",
			true,
			[]utils.SegConfig{
				{ContentID: -1, Hostname: "host", Role: "m"},
				{ContentID: 0, Hostname: "host", Role: "p"},
				{ContentID: 0, Hostname: "host", Role: "m"},
				{ContentID: 1, Hostname: "host", Role: "p"},
				{ContentID: 1, Hostname: "host", Role: "m"},
			},
		},
		{"returns all primary segments excluding master for a given host when includeMirrors is false",
			false,
			[]utils.SegConfig{
				{ContentID: 0, Hostname: "host", Role: "p"},
				{ContentID: 1, Hostname: "host", Role: "p"},
			},
		},
		{"returns all segments excluding master for a given host when includeMirrors is true",
			true,
			[]utils.SegConfig{
				{ContentID: -1, Hostname: "host", Role: "m"},
				{ContentID: 0, Hostname: "host", Role: "p"},
				{ContentID: 0, Hostname: "host", Role: "m"},
				{ContentID: 1, Hostname: "host", Role: "p"},
				{ContentID: 1, Hostname: "host", Role: "m"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := cluster.SegmentsOnExcludingMaster("host", c.includeMirrors)
			if err != nil {
				t.Fatalf("got unexpected error: %+v", err)
			}

			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("got %#v want %#v", actual, c.expected)
			}
		})
	}
}

func TestClusterFromDB(t *testing.T) {
	testStateDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("got error when creating tempdir: %+v", err)
	}

	testhelper.SetupTestLogger()

	defer func() {
		os.RemoveAll(testStateDir)
	}()

	t.Run("returns an error if connection fails", func(t *testing.T) {
		connErr := errors.New("connection failed")
		conn := dbconn.NewDBConnFromEnvironment("testdb")
		conn.Driver = testhelper.TestDriver{ErrToReturn: connErr}

		actualCluster, err := utils.ClusterFromDB(conn, "")

		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if actualCluster != nil {
			t.Errorf("Expected cluster to be nil, but got %#v", actualCluster)
		}
		if !strings.Contains(err.Error(), connErr.Error()) {
			t.Errorf("Expected error: %+v got: %+v", connErr.Error(), err.Error())
		}
	})

	t.Run("returns an error if the segment configuration query fails", func(t *testing.T) {
		conn, mock := testutils.CreateMockDBConn()
		testhelper.ExpectVersionQuery(mock, "5.3.4")

		queryErr := errors.New("failed to get segment configuration")
		mock.ExpectQuery("SELECT .* FROM gp_segment_configuration").WillReturnError(queryErr)

		actualCluster, err := utils.ClusterFromDB(conn, "")

		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
		if actualCluster != nil {
			t.Errorf("Expected cluster to be nil, but got %#v", actualCluster)
		}
		if !strings.Contains(err.Error(), queryErr.Error()) {
			t.Errorf("Expected error: %+v got: %+v", queryErr.Error(), err.Error())
		}
	})

	t.Run("populates a cluster using DB information", func(t *testing.T) {
		conn, mock := testutils.CreateMockDBConn()

		testhelper.ExpectVersionQuery(mock, "5.3.4")
		mock.ExpectQuery("SELECT .* FROM gp_segment_configuration").WillReturnRows(testutils.MockSegmentConfiguration())

		binDir := "/usr/local/gpdb/bin"

		actualCluster, err := utils.ClusterFromDB(conn, binDir)
		if err != nil {
			t.Errorf("got unexpected error: %+v", err)
		}

		expectedCluster := testutils.MockCluster()
		expectedCluster.Version = dbconn.NewVersion("5.3.4")
		expectedCluster.BinDir = binDir

		if !reflect.DeepEqual(actualCluster, expectedCluster) {
			t.Errorf("expected: %#v got: %#v", expectedCluster, actualCluster)
		}
	})
}
