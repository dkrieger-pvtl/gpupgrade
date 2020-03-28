package agent

import (
	"context"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/hub"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) RenameDirectories(ctx context.Context, in *idl.RenameDirectoriesRequest) (*idl.RenameDirectoriesReply, error) {
	gplog.Info("agent received request to rename segment data directories")

	for _, pair := range in.GetPairs() {

		// idempotence here works as follows:
		//  src -> dst:
		//      1). never called, works
		//      2). called before and failed, same as 1) as rename is atomic
		//      4). called before and success, ENOENT
		if err := utils.System.Rename(pair.Src, pair.Dst); err != nil {
			renameErr := hub.RenameError(err)
			if renameErr != nil {
				return &idl.RenameDirectoriesReply{}, renameErr
			}
		}

	}

	return &idl.RenameDirectoriesReply{}, nil
}
