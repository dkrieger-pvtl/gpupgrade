package hub

import (
	"fmt"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/utils"
)

func SwapDataDirectories(config *Config) error {
	swapper := swap{}

	for _, segment := range config.Source.Primaries {
		modifiedSegment, ok := swapper.archiveSource(segment)
		if ok {
			config.Source.Primaries[segment.ContentID] = modifiedSegment
		}
	}

	for _, segment := range config.Target.Primaries {
		modifiedSegment, ok := swapper.promoteTarget(segment)
		if ok {
			config.Target.Primaries[segment.ContentID] = modifiedSegment
		}
	}

	return swapper.errors()
}

func stripUpgradeFromPath(segment utils.SegConfig) string {
	buffer := make([]byte, len(segment.DataDir))

	copy(buffer, segment.DataDir)

	return strings.Replace(string(buffer),
		"_upgrade/",
		"/",
		1)
}

func renameDirectory(originalName, newName string) error {
	gplog.Info(fmt.Sprintf("moving directory %v to %v", originalName, newName))

	return utils.System.Rename(originalName, newName)
}

type swap struct {
	multiErr *multierror.Error
}

func (s *swap) archiveSource(segment utils.SegConfig) (utils.SegConfig, bool) {
	newPath := segment.DataDir + "_old"
	err := renameDirectory(segment.DataDir, newPath)
	s.multiErr = multierror.Append(s.multiErr, err)
	segment.DataDir = newPath
	return segment, nil == err
}

func (s *swap) promoteTarget(segment utils.SegConfig) (utils.SegConfig, bool) {
	newPath := stripUpgradeFromPath(segment)
	err := renameDirectory(segment.DataDir, newPath)
	s.multiErr = multierror.Append(s.multiErr, err)
	segment.DataDir = newPath
	return segment, nil == err
}

func (s *swap) errors() error {
	return s.multiErr.ErrorOrNil()
}
