package hub

import (
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
)

func (s *Server) Finalize(_ *idl.FinalizeRequest, stream idl.CliToHub_FinalizeServer) (err error) {
	st, err := step.Begin(s.StateDir, "finalize", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := st.Finish(); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("finalize: %s", err))
		}
	}()

	st.Run(idl.Substep_FINALIZE_SHUTDOWN_TARGET_CLUSTER, func(streams step.OutStreams) error {
		err := StopCluster(streams, s.Target)

		if err != nil {
			return xerrors.Errorf("failed to stop target cluster: %w", err)
		}

		return nil
	})

	st.Run(idl.Substep_FINALIZE_UPDATE_TARGET_CATALOG_AND_CLUSTER_CONFIG, func(streams step.OutStreams) error {
		return s.UpdateCatalogAndClusterConfig(streams)
	})

	st.Run(idl.Substep_FINALIZE_RENAME_DATA_DIRECTORIES, func(_ step.OutStreams) error {
		return s.RenameDataDirectories()
	})

	st.Run(idl.Substep_FINALIZE_UPDATE_TARGET_CONF_FILES, func(_ step.OutStreams) error {
		return s.UpdateConfFiles()
	})

	st.Run(idl.Substep_FINALIZE_START_TARGET_CLUSTER, func(streams step.OutStreams) error {
		err := StartCluster(streams, s.Target)

		if err != nil {
			return xerrors.Errorf("failed to start target cluster: %w", err)
		}

		return nil
	})

	// This runner runs all commands against the target cluster.
	targetRunner := &greenplumRunner{
		masterPort:          s.Target.MasterPort(),
		masterDataDirectory: s.Target.MasterDataDir(),
		binDir:              s.Target.BinDir,
	}

	// todo: we don't currently have a way to output nothing to the UI when there is no standby.
	// If we did, this check would actually be in `UpgradeStandby`
	if s.Source.HasStandby() {
		st.Run(idl.Substep_FINALIZE_UPGRADE_STANDBY, func(streams step.OutStreams) error {
			// XXX this probably indicates a bad abstraction
			targetRunner.streams = streams

			// TODO: Persist the standby to config.json and update the
			//  target cluster.
			// todo: replace StandbyConfig with the TargetInitializeConfig
			standby := s.TargetInitializeConfig.Standby
			return UpgradeStandby(targetRunner, StandbyConfig{
				Port:          standby.Port,
				Hostname:      standby.Hostname,
				DataDirectory: standby.DataDir,
			})
		})
	}

	// todo: we don't currently have a way to output nothing to the UI when there are no mirrors.
	// If we did, this check would actually be in `UpgradeMirrors`
	if s.Source.HasMirrors() {
		st.Run(idl.Substep_FINALIZE_UPGRADE_MIRRORS, func(streams step.OutStreams) error {
			// XXX this probably indicates a bad abstraction
			targetRunner.streams = streams

			return UpgradeMirrors(s.StateDir, s.Target.MasterPort(), &s.TargetInitializeConfig, targetRunner)
		})
	}

	message := MakeTargetClusterMessage(s.Target)
	if err = stream.Send(message); err != nil {
		return err
	}

	return st.Err()
}
