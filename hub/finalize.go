package hub

import (
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
)

func (s *Server) Finalize(_ *idl.FinalizeRequest, stream idl.CliToHub_FinalizeServer) (err error) {
	agentConnections, err := s.AgentConns()
	if err != nil {
		return err
	}

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

	st.Run(idl.Substep_FINALIZE_SWAP_DATA_DIRECTORIES, func(streams step.OutStreams) error {
		agentBroker := AgentBrokerGRPC{
			agentConnections: agentConnections,
		}

		return SwapDataDirectories(MakeHub(s.Config), &agentBroker)
	})

	st.Run(idl.Substep_FINALIZE_START_TARGET_MASTER, func(streams step.OutStreams) error {
		var cloneOfTarget = *s.Target
		targetMasterConfig := cloneOfTarget.Primaries[-1]
		targetMasterConfig.DataDir = targetMasterConfig.PromotionDataDirectory(s.Source.Primaries[-1])
		cloneOfTarget.Primaries[-1] = targetMasterConfig

		return StartMasterOnly(streams, &cloneOfTarget, false)
	})

	// Once UpdateCatalog && UpdateMasterConf is executed, the port on which the target
	// cluster starts is changed in the catalog and postgresql.conf, however the server config.json target port is
	// still the old port on which the target cluster was initialized.
	// TODO: if any steps needs to connect to the new cluster (that should use new port), we should either
	// write it to the config.json or add some way to identify the state.
	st.Run(idl.Substep_FINALIZE_UPDATE_CATALOG, func(streams step.OutStreams) error {
		return UpdateCatalog(s.Source, s.Target)
	})

	st.Run(idl.Substep_FINALIZE_SHUTDOWN_TARGET_MASTER, func(streams step.OutStreams) error {
		return StopMasterOnly(streams, s.Target, false)
	})

	st.Run(idl.Substep_FINALIZE_UPDATE_POSTGRESQL_CONF, func(streams step.OutStreams) error {
		return UpdateMasterConf(s.Source, s.Target)
	})

	st.Run(idl.Substep_FINALIZE_START_TARGET_CLUSTER, func(streams step.OutStreams) error {
		return StartCluster(streams, s.Target, false)
	})

	return st.Err()
}

func MakeHub(config *Config) Hub {
	var segmentPairsByHost = make(map[string][]SegmentPair)

	for contentId, sourceSegment := range config.Source.Primaries {
		if contentId == -1 {
			continue
		}

		if segmentPairsByHost[sourceSegment.Hostname] == nil {
			segmentPairsByHost[sourceSegment.Hostname] = []SegmentPair{}
		}

		segmentPairsByHost[sourceSegment.Hostname] = append(segmentPairsByHost[sourceSegment.Hostname], SegmentPair{
			source: sourceSegment,
			target: config.Target.Primaries[contentId],
		})
	}

	var configs []Agent
	for hostname, agentSegmentPairs := range segmentPairsByHost {
		configs = append(configs, Agent{
			hostname: hostname,
			pairs:    agentSegmentPairs,
		})
	}

	return Hub{
		sourceMaster: config.Source.Primaries[-1],
		targetMaster: config.Target.Primaries[-1],
		agents:       configs,
	}
}
