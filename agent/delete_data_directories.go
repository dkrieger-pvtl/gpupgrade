package agent

import (
	"context"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) DeleteDirectories(ctx context.Context, in *idl.DeleteDirectoriesRequest) (*idl.DeleteDirectoriesReply, error) {
	gplog.Info("got a request to delete data directories from the hub")

	mErr := &multierror.Error{}

	for _, segDataDir := range in.Datadirs {

		if err := PostgresOrNonExistent(segDataDir); err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}

		err := utils.System.RemoveAll(segDataDir)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	return &idl.DeleteDirectoriesReply{}, mErr.ErrorOrNil()
}
