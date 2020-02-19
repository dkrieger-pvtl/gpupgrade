package hub

import (
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

type AgentSegmentPair struct {
	source utils.SegConfig
	target utils.SegConfig
}

type AgentConfig struct {
	hostname string
	pairs    []AgentSegmentPair
}

type Hub struct {
	sourceMaster utils.SegConfig
	targetMaster utils.SegConfig
	agents       []AgentConfig
}

func SwapDataDirectories(hub Hub, agentBroker AgentBroker) error {
	errors := &multierror.Error{}
	swapper := utils.FilesystemDirectoryFinalizer{MultiErr: errors}

	swapper.Archive(hub.sourceMaster)
	swapper.Promote(hub.targetMaster, hub.sourceMaster)

	for _, agent := range hub.agents {
		// TODO: parallelize
		err := agentBroker.ReconfigureDataDirectories(agent.hostname, makeRenamePairs(agent.pairs))
		multierror.Append(errors, err)
	}

	return errors.ErrorOrNil()
}

func makeRenamePairs(pairs []AgentSegmentPair) []*idl.RenamePair {
	var renamePairs []*idl.RenamePair

	for _, pair := range pairs {
		//// add rename of source to archived
		//renamePairs = append(renamePairs, &idl.RenamePair{
		//	Dst: pair.source.DataDir,
		//	Src: pair.target.DataDir,
		//})

		renamePairs = append(renamePairs, &idl.RenamePair{
			Dst: pair.source.DataDir,
			Src: pair.target.DataDir,
		})
	}

	return renamePairs
}
