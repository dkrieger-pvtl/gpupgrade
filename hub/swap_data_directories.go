package hub

import "github.com/greenplum-db/gpupgrade/utils"

func SwapDataDirectories(config *Config) error {
	swapper := utils.FilesystemDirectoryFinalizer{}

	for _, segment := range config.Source.Primaries {
		swapper.Archive(segment)
	}

	for _, segment := range config.Target.Primaries {
		swapper.Promote(segment)
	}

	return swapper.Errors()
}
