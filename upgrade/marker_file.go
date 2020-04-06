package upgrade

import (
	"fmt"

	"path/filepath"

	"github.com/greenplum-db/gpupgrade/idl"
)

func MarkerFileName(dataDir string, upgradeID ID, kind idl.ClusterType) string {
	return filepath.Join(dataDir, fmt.Sprintf(".gpupgrade_%s_%s", kind, upgradeID))
}
