//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package hub

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/step"
	"golang.org/x/xerrors"
)

//   DIR
//   ├── filespace.txt
//   ├── master
//   │   ├── demoDataDir-1
//   │   │   └── 16385
//   │   │       ├── 1
//   │   │       │   └── GPDB_6_301908232
//   │   │       │       └── 12812
//   │   │       │           └── 16389
//   │   │       └── 12094
//   │   │           ├── 16384
//   │   │           └── PG_VERSION
//   ├── primary1
//   │   └── demoDataDir0
//   │       └── 16385
//   │           ├── 12094
//   │           │   ├── 16384
//   │           │   └── PG_VERSION
//   │           └── 2
//   │               └── GPDB_6_301908232
//   │                   └── 12812
//   │                       └── 16389
//
//  GPDB-5:  DIR/<fsname>/<datadir>/<tablespace_oid>/<database_oid>/<relfilenode>
//  GPDB-6   DIR/<fsname>/<datadir>/<tablespace_oid>/<dboid>/GPDB_6_<catalog_version>/<database_oid>/<relfilenode>
//
//   We use the GPDB-5 tablespace mapping read during Initialize to construct the paths
//         of the tablespaces in 6.  There is a known mapping.
//
// Do we handle temporary and transaction files? not needed
//
//postgres --catalog-version
//Catalog version number:               301908232

type TablespacesOnDBID = map[int][]string

// GetTablespaceMapping returns per-dbid slice of directories of user-defined tablespaces.
func GetTablespaceMapping(in greenplum.Tablespaces) TablespacesOnDBID {
	m := make(TablespacesOnDBID)
	for dbid, segmentTbsp := range in {
		for _, tbspInfo := range segmentTbsp {
			if tbspInfo.IsUserDefined() {
				m[dbid] = append(m[dbid], tbspInfo.Location)
			}
		}
	}
	return m
}

//  GPDB-5:  DIR/<fsname>/<datadir>/<tablespace_oid>/<database_oid>/<relfilenode>
//           /tmp/fs/m3/demoDataDir2/16448
//  GPDB-6   DIR/<fsname>/<datadir>/<tablespace_oid>/<dboid>/GPDB_6_<catalog_version>/<database_oid>/<relfilenode>
func GetGPDB6TablespaceMapping(in5 TablespacesOnDBID, gpversion string) TablespacesOnDBID {
	m := make(TablespacesOnDBID)

	for dbid, tbsps5 := range in5 {
		for _, tbsp5 := range tbsps5 {
			pathel := "GPDB_6_" + gpversion
			tbsp6 := filepath.Join(tbsp5, strconv.Itoa(dbid), pathel)
			m[dbid] = append(m[dbid], tbsp6)
		}
	}
	return m
}

// GetCatalogVersion uses the postgres binary to determine the clusters catalog version
// postgres --catalog-version
//   Catalog version number:               301908232
// TODO: make sure the target datadirs contain our upgradeID
// TODO: consider using pg_controldata instead...it's likely more up to date
// TODO: consider just hardcoding the catalog ID and eliminating this function?
//   The catalog version "never" changes for 6...
func GetCatalogVersion(bindir string) (string, error) {
	path := filepath.Join(bindir, "postgres")

	cmd := exec.Command(path, "--catalog-version")

	// Explicitly clear the child environment.
	cmd.Env = []string{}

	// XXX ...but we make a single exception for now, for LD_LIBRARY_PATH, to
	// work around pervasive problems with RPATH settings in our Postgres
	// extension modules.
	if path, ok := os.LookupEnv("LD_LIBRARY_PATH"); ok {
		cmd.Env = append(cmd.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", path))
	}

	stream := &step.BufferedStreams{}
	cmd.Stdout = stream.Stdout()
	cmd.Stderr = stream.Stderr()

	err := cmd.Run()
	if err != nil {
		return "", xerrors.Errorf("could not determine catalog version: %w", err)
	}

	s := strings.Split(stream.StdoutBuf.String(), ":")
	if len(s) != 2 {
		return "", xerrors.Errorf("unexpected catalog version string: %s", stream.StdoutBuf.String())
	}
	key := strings.TrimSpace(s[0])
	if key != "Catalog version number" {
		return "", xerrors.Errorf("unexpected catalog version key: %s", key)
	}
	value := strings.TrimSpace(s[1])
	if len(value) != 9 || !strings.HasPrefix(value, "30") {
		return "", xerrors.Errorf("unexpected catalog version: %s", value)
	}

	return value, nil

}

func DeleteTableSpaceDirectories(in TablespacesOnDBID) error {
	return nil
}
