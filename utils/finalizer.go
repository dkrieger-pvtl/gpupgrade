package utils

import (
	"strings"
)

type DataDirFinalizer struct {
}

func (finalizer *DataDirFinalizer) Archive(segment SegConfig) SegConfig {
	newPath := segment.DataDir + "_old"
	segment.DataDir = newPath
	return segment
}

func (finalizer *DataDirFinalizer) Promote(segment SegConfig) SegConfig {
	buffer := make([]byte, len(segment.DataDir))

	copy(buffer, segment.DataDir)

	newPath := strings.Replace(string(buffer),
		"_upgrade/",
		"/",
		1)

	segment.DataDir = newPath

	return segment
}
