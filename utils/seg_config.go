package utils

func (segment SegConfig) ArchiveDataDirectory() string {
	return segment.DataDir + "_old"
}

func (segment SegConfig) PromotionDataDirectory(sourceSegment SegConfig) string {
	return sourceSegment.DataDir
}
