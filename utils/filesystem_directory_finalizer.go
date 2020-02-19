package utils

import (
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
)

type FilesystemDirectoryFinalizer struct {
	MultiErr           *multierror.Error
	directoryFinalizer *DataDirFinalizer
}

func (s *FilesystemDirectoryFinalizer) Archive(segment SegConfig) {
	archivedSegment := s.directoryFinalizer.Archive(segment)
	err := renameDirectory(segment.DataDir, archivedSegment.DataDir)
	s.MultiErr = multierror.Append(s.MultiErr, err)
}

func (s *FilesystemDirectoryFinalizer) Promote(segment SegConfig, sourceSegment SegConfig) {
	promotedSegment := s.directoryFinalizer.Promote(segment, sourceSegment)
	err := renameDirectory(segment.DataDir, promotedSegment.DataDir)
	s.MultiErr = multierror.Append(s.MultiErr, err)
}

func renameDirectory(originalName, newName string) error {
	gplog.Info(fmt.Sprintf("moving directory %v to %v", originalName, newName))

	return System.Rename(originalName, newName)
}
