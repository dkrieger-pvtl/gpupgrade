// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/upgrade"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) Revert(_ *idl.RevertRequest, stream idl.CliToHub_RevertServer) (err error) {
	st, err := step.Begin(s.StateDir, "revert", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := st.Finish(); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("revert: %s", err))
		}
	}()

	// TODO: get design input on this message
	if !s.Source.HasAllMirrorsAndStandby() {
		return errors.New("gpupgrade revert is only supported for clusters with all mirrors and a standby")
	}

	// Since revert needs to work at any point, and stop is not yet idempotent
	// check if the cluster is running before stopping.
	// TODO: This will fail if the target does not exist which can occur when
	//  initialize fails part way through and does not create the target cluster.
	running, err := s.Target.IsMasterRunning(st.Streams())
	if err != nil {
		return err
	}

	if running {
		st.Run(idl.Substep_SHUTDOWN_TARGET_CLUSTER, func(streams step.OutStreams) error {
			if err := s.Target.Stop(streams); err != nil {
				return xerrors.Errorf("stopping target cluster: %w", err)
			}
			return nil
		})
	}

	// This substep needs to be conditionalized on the status of the upgrade; it depends on how far execute
	//   ran; see "Reverting to old cluster" in https://www.postgresql.org/docs/9.4/pgupgrade.html.
	// For now, we only handle revert after initialize or execute fully succeeds.
	// We do this in link and copy mode due to the recoverseg bug
	// TODO: only run if we are at the end of execute
	st.Run(idl.Substep_RESTORE_SOURCE_MASTER_AND_PRIMARIES, func(stream step.OutStreams) error {
		return RestoreMasterAndPrimaries(stream, s.Source, s.agentConns)
	})

	// we no longer need the target cluster after revent.
	// Note deleting hard-linked files just lowers the refcount from 2 to 1.
	if len(s.Config.Target.Primaries) > 0 {
		st.Run(idl.Substep_DELETE_PRIMARY_DATADIRS, func(_ step.OutStreams) error {
			return DeletePrimaryDataDirectories(s.agentConns, s.Config.Target)
		})

		st.Run(idl.Substep_DELETE_MASTER_DATADIR, func(streams step.OutStreams) error {
			datadir := s.Config.Target.MasterDataDir()
			hostname := s.Config.Target.MasterHostname()

			return upgrade.DeleteDirectories([]string{datadir}, upgrade.PostgresFiles, hostname, streams)
		})

		// see comments below
		st.Run(idl.Substep_DELETE_TABLESPACE_DATADIRS, func(streams step.OutStreams) error {
			gpdb5 := GetTablespaceMapping(s.Tablespaces)
			// TODO: validate this set against the cluster config
			GetCatalogVersion(s.Target.BinDir)
			GetGPDB6TablespaceMapping(gpdb5)
			return nil
		})
	}

	st.Run(idl.Substep_ARCHIVE_LOG_DIRECTORIES, func(_ step.OutStreams) error {
		// Archive log directory on master
		oldDir, err := utils.GetLogDir()
		if err != nil {
			return err
		}
		newDir := filepath.Join(filepath.Dir(oldDir), utils.GetArchiveDirectoryName(time.Now()))
		if err = utils.System.Rename(oldDir, newDir); err != nil {
			if utils.System.IsNotExist(err) {
				gplog.Debug("log directory %s not archived, possibly due to multi-host environment. %+v", newDir, err)
			}
		}

		return ArchiveSegmentLogDirectories(s.agentConns, s.Config.Target.MasterHostname(), newDir)
	})

	st.Run(idl.Substep_DELETE_SEGMENT_STATEDIRS, func(_ step.OutStreams) error {
		return DeleteStateDirectories(s.agentConns, s.Source.MasterHostname())
	})

	// Since revert needs to work at any point, and start is not yet idempotent
	// check if the cluster is not running before starting.
	running, err = s.Source.IsMasterRunning(st.Streams())
	if err != nil {
		return err
	}

	if !running {
		st.Run(idl.Substep_START_SOURCE_CLUSTER, func(streams step.OutStreams) error {
			if err := s.Source.Start(streams); err != nil {
				return xerrors.Errorf("starting source cluster: %w", err)
			}
			return nil
		})
	}

	return st.Err()
}

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

func GetGPDB6TablespaceMapping(in TablespacesOnDBID) TablespacesOnDBID {
	m := make(TablespacesOnDBID)

	return m
}

// GetCatalogVersion uses the postgres binary to determine the clusters catalog version
// postgres --catalog-version
//   Catalog version number:               301908232
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
