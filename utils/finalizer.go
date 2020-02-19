package utils

type DataDirFinalizer struct {
}

func (finalizer *DataDirFinalizer) Archive(segment SegConfig) SegConfig {
	newPath := segment.DataDir + "_old"
	segment.DataDir = newPath
	return segment
}

func (finalizer *DataDirFinalizer) Promote(segment SegConfig, sourceSegment SegConfig) SegConfig {
	segment.DataDir = sourceSegment.DataDir
	return segment
}
