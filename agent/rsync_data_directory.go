package agent

// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

import (
	"context"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils/rsync"
)

// TODO: we launch rsync on the mirror and push to the primary; is this the right order?
func (s *Server) RsyncDataDirectory(ctx context.Context, in *idl.RsyncDataDirectoryRequest) (*idl.RsyncDataDirectoryReply, error) {
	gplog.Info("agent received request to rsync data directory.")

	var wg sync.WaitGroup
	errs := make(chan error, len(in.Pairs))

	// todo: make sure these are, in fact, data dirs.

	for _, pair := range in.Pairs {
		pair := pair
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- rsync.RsyncWithoutStream(pair.Src, pair.DstHostname, pair.Dst, in.Options, in.Exclude)
		}()
	}

	wg.Wait()
	close(errs)

	var mErr *multierror.Error
	for err := range errs {
		mErr = multierror.Append(mErr, err)
	}

	return &idl.RsyncDataDirectoryReply{}, mErr.ErrorOrNil()
}
