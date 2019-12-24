package hub

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (h *Hub) Execute(request *idl.ExecuteRequest, stream idl.CliToHub_ExecuteServer) (err error) {
	// Create a log file to contain execute output.
	log, err := utils.System.OpenFile(
		filepath.Join(utils.GetStateDir(), "execute.log"),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := log.Close(); closeErr != nil {
			err = multierror.Append(err,
				xerrors.Errorf("failed to close execute log: %w", closeErr))
		}
	}()

	executeStream := newMultiplexedStream(stream, log)

	_, err = log.WriteString("\nExecute in progress.\n")
	if err != nil {
		return xerrors.Errorf("failed writing to execute log: %w", err)
	}

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.UPGRADE_MASTER, idl.UpgradeSteps_UPGRADE_MASTER)
	err = h.Substep(executeStream, upgradestatus.UPGRADE_MASTER,
		func(streams OutStreams) error {
			return h.UpgradeMaster(streams, false)
		})
	if err != nil {
		return err
	}

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.COPY_MASTER, idl.UpgradeSteps_COPY_MASTER)
	err = h.Substep(executeStream, upgradestatus.COPY_MASTER, h.CopyMasterDataDir)
	if err != nil {
		return err
	}

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.UPGRADE_PRIMARIES, idl.UpgradeSteps_UPGRADE_PRIMARIES)
	err = h.Substep(executeStream, upgradestatus.UPGRADE_PRIMARIES,
		func(_ OutStreams) error {
			return h.ConvertPrimaries(false)
		})
	if err != nil {
		return err
	}

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.START_TARGET_CLUSTER, idl.UpgradeSteps_START_TARGET_CLUSTER)
	err = h.Substep(executeStream, upgradestatus.START_TARGET_CLUSTER,
		func(streams OutStreams) error {
			return StartCluster(streams, h.target)
		})
	return err
}
