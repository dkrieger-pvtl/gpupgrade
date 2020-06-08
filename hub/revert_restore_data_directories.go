// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package hub

import (
	"context"
	"os"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils/rsync"
)

func RestoreMasterAndPrimaries(stream step.OutStreams, source *greenplum.Cluster, agentConns []*Connection) error {
	exclude := []string{
		"pg_hba.conf", "postmaster.opts", "postgresql.auto.conf", "internal.auto.conf", "gp_dbid",
		"postgresql.conf", "backup_label.old", "postmaster.pid", "recovery.conf",
	}
	options := []string{"--archive", "--compress", "--stats"}

	if !source.HasAllMirrorsAndStandby() {
		return errors.New("gpupgrade revert is only supported for clusters with all mirrors and a standby")
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(agentConns)+1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- rsync.RsyncWithStream(source.Mirrors[-1].DataDir, source.MasterHostname(), source.MasterDataDir(), options, exclude, stream)
	}()

	for _, conn := range agentConns {

		conn := conn

		wg.Add(1)
		go func() {
			defer wg.Done()

			var pairs []*idl.RsyncDataDirPair
			mirrors := source.SelectSegments(func(seg *greenplum.SegConfig) bool {
				return seg.IsOnHost(conn.Hostname) && !seg.IsStandby() && seg.IsMirror()
			})
			if len(mirrors) == 0 {
				return
			}

			for _, mirror := range mirrors {
				pair := &idl.RsyncDataDirPair{
					Src:         mirror.DataDir + string(os.PathSeparator), // the trailing slash is critical for rsync
					DstHostname: source.Primaries[mirror.ContentID].Hostname,
					Dst:         source.Primaries[mirror.ContentID].DataDir,
				}
				pairs = append(pairs, pair)
			}

			req := &idl.RsyncDataDirectoryRequest{
				Options: options,
				Exclude: exclude,
				Pairs:   pairs,
			}

			_, err := conn.AgentClient.RsyncDataDirectory(context.Background(), req)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	var multiErr *multierror.Error
	for err := range errs {
		multiErr = multierror.Append(multiErr, err)
	}
	return multiErr.ErrorOrNil()

}
