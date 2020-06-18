package hub_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/hub"
)

func TestGetTablespaceMapping(t *testing.T) {
	tbps := greenplum.Tablespaces{
		1: {
			1663: greenplum.TablespaceInfo{
				Location:    "/qddir/datadir-0",
				UserDefined: 0,
			},
			17000: greenplum.TablespaceInfo{
				Location:    "/mount/fs/17000",
				UserDefined: 1,
			},
		},
		2: {
			1663: greenplum.TablespaceInfo{
				Location:    "/tmp/master/gpseg-1",
				UserDefined: 0,
			},
		},
	}

	expected := hub.TablespacesOnDBID{
		1: []string{"/mount/fs/17000"},
	}
	actual := hub.GetTablespaceMapping(tbps)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %#v, expected %#v", actual, expected)
	}
}

func TestGetCatalogVersion(t *testing.T) {
	value, err := hub.GetCatalogVersion("/usr/local/gpdb6/bin")
	if err != nil {
		t.Errorf("opps: %#v", err)
	}
	fmt.Println("catalog version:", value, "NUM", len(value))
}
