package agent

import (
	"context"

	"github.com/greenplum-db/gpupgrade/utils"

	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/idl"
)

// CreateFiles adds the requested absolute paths.  If a given path already exists, it does not
//  indicate an error.
func (s *Server) CreateFiles(ctx context.Context, in *idl.CreateFilesRequest) (*idl.CreateFilesReply, error) {
	gplog.Info("agent received request to mark segment data directories")

	var mErr *multierror.Error

	for _, file := range in.Files {
		if err := utils.AddEmptyFileIdempotent(file); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	return &idl.CreateFilesReply{}, mErr.ErrorOrNil()
}
