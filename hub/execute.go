package hub

import (
	"fmt"

	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
)

func (h *Hub) Execute(request *idl.ExecuteRequest, stream idl.CliToHub_ExecuteServer) (err error) {
	substeps, err := BeginStep(h.conf.StateDir, "execute", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := FinishStep(substeps); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("execute: %s", err))
		}
	}()

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.UPGRADE_MASTER, idl.UpgradeSteps_UPGRADE_MASTER)
	substeps.Run(idl.UpgradeSteps_UPGRADE_MASTER, func(streams OutStreams) error {
		return h.UpgradeMaster(streams, false)
	})

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.COPY_MASTER, idl.UpgradeSteps_COPY_MASTER)
	substeps.Run(idl.UpgradeSteps_COPY_MASTER, func(streams OutStreams) error {
		return h.CopyMasterDataDir(streams)
	})

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.UPGRADE_PRIMARIES, idl.UpgradeSteps_UPGRADE_PRIMARIES)
	substeps.Run(idl.UpgradeSteps_UPGRADE_PRIMARIES, func(_ OutStreams) error {
		return h.ConvertPrimaries(false)
	})

	h.checklist.(*upgradestatus.ChecklistManager).AddWritableStep(upgradestatus.SHUTDOWN_TARGET_CLUSTER, idl.UpgradeSteps_START_TARGET_CLUSTER)
	substeps.Run(idl.UpgradeSteps_START_TARGET_CLUSTER, func(streams OutStreams) error {
		return StartCluster(streams, h.target)
	})

	return Err(substeps)
}
