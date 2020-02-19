package agent

import (
	"context"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) ReconfigureDataDirectories(ctx context.Context, in *idl.ReconfigureDataDirRequest) (*idl.ReconfigureDataDirReply, error) {
	gplog.Info("got a request to move segment data directories from the hub")

	for _, pair := range in.GetPair() {
		err := utils.System.Rename(pair.Src, pair.Dst)
		if err != nil {
			return &idl.ReconfigureDataDirReply{}, err
		}
	}

	return &idl.ReconfigureDataDirReply{}, nil
}
