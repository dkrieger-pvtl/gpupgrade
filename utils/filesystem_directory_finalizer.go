package utils

import (
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
)

type FilesystemDirectoryFinalizer struct {
	multiErr           *multierror.Error
	directoryFinalizer *DataDirFinalizer
}

func (s *FilesystemDirectoryFinalizer) Archive(segment SegConfig) (SegConfig, bool) {
	archivedSegment := s.directoryFinalizer.Archive(segment)
	err := renameDirectory(segment.DataDir, archivedSegment.DataDir)
	s.multiErr = multierror.Append(s.multiErr, err)
	return segment, nil == err
}

func (s *FilesystemDirectoryFinalizer) Promote(segment SegConfig) (SegConfig, bool) {
	promotedSegment := s.directoryFinalizer.Promote(segment)

	err := renameDirectory(segment.DataDir, promotedSegment.DataDir)
	s.multiErr = multierror.Append(s.multiErr, err)
	return segment, nil == err
}

func (s *FilesystemDirectoryFinalizer) Errors() error {
	return s.multiErr.ErrorOrNil()
}

func renameDirectory(originalName, newName string) error {
	gplog.Info(fmt.Sprintf("moving directory %v to %v", originalName, newName))

	return System.Rename(originalName, newName)
}
