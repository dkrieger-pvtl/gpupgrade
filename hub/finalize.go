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
	st, err := BeginStep(s.StateDir, "finalize", stream)
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
		return StopCluster(streams, s.Target, false)
	})

	st.Run(idl.Substep_FINALIZE_START_TARGET_MASTER, func(streams step.OutStreams) error {
		return StartMasterOnly(streams, s.Target, false)
	})

	// Once UpdateCatalogWithPortInformation && UpdateMasterPostgresqlConf is executed, the port on which the target
	// cluster starts is changed in the catalog and postgresql.conf, however the server config.json target port is
	// still the old port on which the target cluster was initialized.
	// TODO: if any steps needs to connect to the new cluster (that should use new port), we should either
	// write it to the config.json or add some way to identify the state.
	st.Run(idl.Substep_FINALIZE_UPDATE_CATALOG_WITH_PORT, func(streams step.OutStreams) error {
		return UpdateCatalogWithPortInformation(s.Source, s.Target)
	})

	st.Run(idl.Substep_FINALIZE_SHUTDOWN_TARGET_MASTER, func(streams step.OutStreams) error {
		return StopMasterOnly(streams, s.Target, false)
	})

	st.Run(idl.Substep_FINALIZE_UPDATE_POSTGRESQL_CONF, func(streams step.OutStreams) error {
		return UpdateMasterPostgresqlConf(s.Source, s.Target)
	})

	st.Run(idl.Substep_FINALIZE_UPDATE_TARGET_CATALOG_WITH_DATADIRS, func(streams step.OutStreams) error {
		// /data/qddir_upgrade/demoDataDir-1 -> datadirs/qddir/demoDataDir-1
		return ChangeDataDirsInCatalog(s.Target, upgradeDataDir, ModifySegmentCatalog)
	})

	//st.Run(idl.Substep_FINALIZE_RENAME_TARGET_MASTER_DATADIR, func(streams step.OutStreams) error {
	//	// /data/qddir_upgrade/demoDataDir-1 -> datadirs/qddir/demoDataDir-1
	//	src := upgradeDataDir(s.Target.MasterDataDir())
	//	dst := s.Target.MasterDataDir()
	//	err := utils.System.Rename(src, dst)
	//	if err != nil {
	//		return xerrors.Errorf("renaming target cluster master datadir from: %s to: %s", src, dst, err)
	//	}
	//	return nil
	//})
	//
	//st.Run(idl.Substep_FINALIZE_RENAME_TARGET_SEG_DATADIRS, func(streams step.OutStreams) error {
	//	// /data/dbfast1_upgrade/demoDataDir0 -> datadirs/dbfast1/demoDataDir0
	//	err := renameSegmentDataDirsOnDisk(s.agentConns, s.Target, upgradeDataDir, noop)
	//	if err != nil {
	//		return xerrors.Errorf("renaming target directories: %w")
	//	}
	//	return nil
	//})
	//
	//st.Run(idl.Substep_FINALIZE_RENAME_SOURCE_MASTER_DATADIR, func(streams step.OutStreams) error {
	//	// /data/qddir/demoDataDir-1 -> datadirs/qddir_old/demoDataDir-1
	//	src := s.Source.MasterDataDir()
	//	dst := oldDataDir(s.Source.MasterDataDir())
	//	err = utils.System.Rename(src, dst)
	//	if err != nil {
	//		return xerrors.Errorf("renaming source cluster master datadir from: %s to: %s", src, dst, err)
	//	}
	//	return nil
	//})
	//
	//st.Run(idl.Substep_FINALIZE_RENAME_SOURCE_SEG_DATADIRS, func(streams step.OutStreams) error {
	//	// /data/dbfast1/demoDataDir0 -> datadirs/dbfast1_old/demoDataDir0
	//	err = renameSegmentDataDirsOnDisk(s.agentConns, s.Source, noop, oldDataDir)
	//	if err != nil {
	//		return xerrors.Errorf("renaming source directories: %w")
	//	}
	//	return nil
	//})

	st.Run(idl.Substep_FINALIZE_UPDATE_GPPERFMON_CONF, func(streams step.OutStreams) error {
		err = updateGpperfmonConf(upgradeDataDir(s.Target.MasterDataDir()), s.Target.MasterDataDir())
		if err != nil {
			return xerrors.Errorf("updating target cluster gpperfmon.conf: %w")
		}

		err := updateGpperfmonConf(s.Source.MasterDataDir(), oldDataDir(s.Source.MasterDataDir()))
		if err != nil {
			return xerrors.Errorf("updating source cluster gpperfmon.conf: %w")
		}

		return nil
	})

	st.Run(idl.Substep_FINALIZE_START_TARGET_CLUSTER, func(streams step.OutStreams) error {
		return StartCluster(streams, s.Target, false)
	})

	return st.Err()
}

var noop = func(path string) string {
	return path
}
