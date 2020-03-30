package agent

import (
	"context"

	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/hub"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) RenameDirectories(ctx context.Context, in *idl.RenameDirectoriesRequest) (*idl.RenameDirectoriesReply, error) {
	gplog.Info("agent received request to rename segment data directories")

	mErr := &multierror.Error{}

	for _, pair := range in.GetPairs() {

		if err := PostgresOrNonExistent(pair.Src); err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}
		if err := PostgresOrNonExistent(pair.Dst); err != nil {
			mErr = multierror.Append(mErr, err)
			continue
		}

		if err := utils.System.Rename(pair.Src, pair.Dst); err != nil {
			if !hub.IsRenameErrorIdempotent(err) {
				return &idl.RenameDirectoriesReply{}, err
			}
		}

	}

	return &idl.RenameDirectoriesReply{}, mErr.ErrorOrNil()
}
