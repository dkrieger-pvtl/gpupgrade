package hub

import (
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

type SegmentPair struct {
	source utils.SegConfig
	target utils.SegConfig
}

type Agent struct {
	hostname string
	pairs    []SegmentPair
}

type Hub struct {
	sourceMaster utils.SegConfig
	targetMaster utils.SegConfig
	agents       []Agent
}

type finalizer struct {
	err         *multierror.Error
	agentBroker AgentBroker
}

func (f *finalizer) archive(segment utils.SegConfig) {
	err := renameDirectory(segment.DataDir, segment.ArchiveDataDirectory())
	f.err = multierror.Append(f.err, err)
}

func (f *finalizer) promote(segment utils.SegConfig, sourceSegment utils.SegConfig) {
	err := renameDirectory(segment.DataDir, segment.PromotionDataDirectory(sourceSegment))
	f.err = multierror.Append(f.err, err)
}

func SwapDataDirectories(hub Hub, agentBroker AgentBroker) error {
	swapper := finalizer{
		agentBroker: agentBroker,
	}

	swapper.archive(hub.sourceMaster)
	swapper.promote(hub.targetMaster, hub.sourceMaster)
	swapper.swapDirectoriesOnAgents(hub.agents)

	return swapper.Errors()
}

func (f *finalizer) swapDirectoriesOnAgents(agents []Agent) {
	result := make(chan error, len(agents))

	gplog.Info("Working with %d agents", len(agents))
	gplog.Info("Working with agents: %+v", agents)

	for _, agent := range agents {
		agent := agent

		//TODO: make this use of agentBroker multi-thread safe inherently.
		go func() {
			result <- f.agentBroker.ReconfigureDataDirectories(agent.hostname,
				makeRenamePairs(agent.pairs))
		}()
	}

	for range agents {
		multierror.Append(f.err, <-result)
	}
}

func (f *finalizer) Errors() error {
	return f.err.ErrorOrNil()
}

func makeRenamePairs(pairs []SegmentPair) []*idl.RenamePair {
	var renamePairs []*idl.RenamePair

	for _, pair := range pairs {
		renamePairs = append(renamePairs, &idl.RenamePair{
			Src: pair.source.DataDir,
			Dst: pair.source.ArchiveDataDirectory(),
		})

		renamePairs = append(renamePairs, &idl.RenamePair{
			Src: pair.target.DataDir,
			Dst: pair.target.PromotionDataDirectory(pair.source),
		})
	}

	return renamePairs
}

func renameDirectory(originalName, newName string) error {
	gplog.Info(fmt.Sprintf("moving directory %v to %v", originalName, newName))

	return utils.System.Rename(originalName, newName)
}
