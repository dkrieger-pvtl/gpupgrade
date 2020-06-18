//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

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
				Location:    "/tmp/fs/m/demoDataDir2/16448",
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
		1: []string{"/tmp/fs/m/demoDataDir2/16448"},
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

func TestGetGPDB6TablespaceMapping(t *testing.T) {
	in := hub.TablespacesOnDBID{
		1: []string{"/tmp/fs/m/demoDataDir2/16448", "/tmp/fs/m/demoDataDir2/20000"},
		2: []string{"/tmp/fs/p1/demoDataDir2/16448", "/tmp/fs/p1/demoDataDir2/20000"},
	}
	expected := hub.TablespacesOnDBID{
		1: []string{"/tmp/fs/m/demoDataDir2/16448/1/GPDB_6_301908232", "/tmp/fs/m/demoDataDir2/20000/1/GPDB_6_301908232"},
		2: []string{"/tmp/fs/p1/demoDataDir2/16448/2/GPDB_6_301908232", "/tmp/fs/p1/demoDataDir2/20000/2/GPDB_6_301908232"},
	}

	actual := hub.GetGPDB6TablespaceMapping(in, "301908232")

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %#v, expected %#v", actual, expected)
	}
}
