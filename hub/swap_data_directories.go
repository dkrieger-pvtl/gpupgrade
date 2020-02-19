package hub

import "github.com/greenplum-db/gpupgrade/utils"

func SwapDataDirectories(config *Config) error {
	swapper := utils.FilesystemDirectoryFinalizer{}

	sourceMaster := config.Source.Primaries[-1]
	swapper.Archive(sourceMaster)

	targetMaster := config.Target.Primaries[-1]
	swapper.Promote(targetMaster, sourceMaster)

	return swapper.Errors()
}
