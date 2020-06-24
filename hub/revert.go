// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/upgrade"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) Revert(_ *idl.RevertRequest, stream idl.CliToHub_RevertServer) (err error) {
	st, err := step.Begin(s.StateDir, idl.Step_REVERT, stream)
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

	if !s.Source.HasAllMirrorsAndStandby() {
		return errors.New("Source cluster does not have mirrors and/or standby. Cannot restore source cluster. Please contact support.")
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

	// Restoring the source master and primaries is only needed if upgrading the
	// primaries had started.
	// TODO: For now we use if the source master is not running to determine this.
	running, err = s.Source.IsMasterRunning(st.Streams())
	if err != nil {
		return err
	}

	if !running {
		// Restoring the master and primaries is needed in copy mode due to an issue
		// in 5X where the source cluster is left in a bad state after execute. This
		// is because running pg_upgrade on a primary results in a checkpoint that
		// does not get replicated on the mirror. Thus, when the mirror is started
		// it panics and a gprecoverseg or rsync is needed.
		st.Run(idl.Substep_RESTORE_SOURCE_MASTER_AND_PRIMARIES, func(stream step.OutStreams) error {
			return RestoreMasterAndPrimaries(stream, s.agentConns, s.Source)
		})
	}

	if len(s.Config.Target.Primaries) > 0 {
		st.Run(idl.Substep_DELETE_PRIMARY_DATADIRS, func(_ step.OutStreams) error {
			return DeletePrimaryDataDirectories(s.agentConns, s.Config.Target)
		})

		st.Run(idl.Substep_DELETE_MASTER_DATADIR, func(streams step.OutStreams) error {
			datadir := s.Config.Target.MasterDataDir()
			hostname := s.Config.Target.MasterHostname()

			return upgrade.DeleteDirectories([]string{datadir}, upgrade.PostgresFiles, hostname, streams)
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
