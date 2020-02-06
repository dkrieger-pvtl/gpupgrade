package hub

import (
	"fmt"
	"os/exec"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
)

func (h *Hub) Execute(request *idl.ExecuteRequest, stream idl.CliToHub_ExecuteServer) (err error) {
	s, err := BeginStep(h.StateDir, "execute", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := s.Finish(); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("execute: %s", err))
		}
	}()

	s.Run(idl.Substep_SHUTDOWN_SOURCE_CLUSTER, func(stream step.OutStreams) error {
		return StopCluster(stream, h.Source)
	})

	s.Run(idl.Substep_UPGRADE_MASTER, func(streams step.OutStreams) error {
		stateDir := h.StateDir
		return UpgradeMaster(h.Source, h.Target, stateDir, streams, false, h.UseLinkMode)
	})

	s.Run(idl.Substep_COPY_MASTER, h.CopyMasterDataDir)

	s.Run(idl.Substep_UPGRADE_PRIMARIES, func(_ step.OutStreams) error {
		return h.ConvertPrimaries(false)
	})

	{
		// TODO: this is only needed on a 5X source cluster
		s.Run(idl.Substep_START_SOURCE_CLUSTER, func(stream step.OutStreams) error {
			err := StartCluster(stream, h.Source)
			if _, ok := err.(*exec.ExitError); ok {
				gplog.Info("exit error on start, hopefully due to mirrors being down")
				return nil
			}
			return err
		})

		s.Run(idl.Substep_SOURCE_RECOVERSEG, func(stream step.OutStreams) error {
			return Recoverseg(stream, h.Source)
		})

		s.AlwaysRun(idl.Substep_SHUTDOWN_SOURCE_CLUSTER, func(stream step.OutStreams) error {
			return StopCluster(stream, h.Source)
		})
	}

	s.Run(idl.Substep_START_TARGET_CLUSTER, func(streams step.OutStreams) error {
		return StartCluster(streams, h.Target)
	})

	return s.Err()
}
