package agent

import (
	"os"
	"path/filepath"

	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/hashicorp/go-multierror"
)

var PostgresFiles = [...]string{"postgresql.conf", "PG_VERSION"}

func PostgresOrNonExistent(dataDir string) error {
	mErr := &multierror.Error{}

	if _, err := utils.System.Stat(dataDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		mErr = multierror.Append(mErr, err)
		return mErr.ErrorOrNil()
	}

	for _, fileName := range PostgresFiles {
		filePath := filepath.Join(dataDir, fileName)
		_, err := utils.System.Stat(filePath)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	return mErr.ErrorOrNil()
}
