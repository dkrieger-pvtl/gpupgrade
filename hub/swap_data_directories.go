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
	MultiErr    *multierror.Error
	agentBroker AgentBroker
}

func (f *finalizer) archive(segment utils.SegConfig) {
	archivedDataDirectory := segment.ArchiveDataDirectory()
	err := renameDirectory(segment.DataDir, archivedDataDirectory)
	f.MultiErr = multierror.Append(f.MultiErr, err)
}

func (f *finalizer) promote(segment utils.SegConfig, sourceSegment utils.SegConfig) {
	promotedDataDir := segment.PromotionDataDirectory(sourceSegment)
	err := renameDirectory(segment.DataDir, promotedDataDir)
	f.MultiErr = multierror.Append(f.MultiErr, err)
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
	for _, agent := range agents {
		err := f.agentBroker.ReconfigureDataDirectories(agent.hostname,
			makeRenamePairs(agent.pairs))
		multierror.Append(f.MultiErr, err)
	}
}

func (f *finalizer) Errors() error {
	return f.MultiErr.ErrorOrNil()
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
