package hub

import (
	"context"
	"sort"
	"sync"

	"github.com/greenplum-db/gpupgrade/utils"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/greenplum"
	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/idl"

	"github.com/greenplum-db/gpupgrade/upgrade"
)

type MarkMap = map[string][]string

func (s *Server) AddMarkerFiles() error {
	agentConns, err := s.AgentConns()
	if err != nil {
		return xerrors.Errorf("cannot collect agents for file renaming: %w", err)
	}
	return AddMarkerFiles(s.Source, s.Target, s.UpgradeID, agentConns)
}

func AddMarkerFiles(source, target *greenplum.Cluster, upgradeID upgrade.ID, agentConns []*Connection) error {

	file := upgrade.MarkerFileName(source.Primaries[-1].DataDir, upgradeID, idl.ClusterType_SOURCE)
	if err := utils.AddEmptyFileIdempotent(file); err != nil {
		return xerrors.Errorf("adding marker source master: %w", err)
	}

	file = upgrade.MarkerFileName(target.Primaries[-1].DataDir, upgradeID, idl.ClusterType_TARGET)
	if err := utils.AddEmptyFileIdempotent(file); err != nil {
		return xerrors.Errorf("adding marker target master: %w", err)
	}

	sourceDirs := getNonMasterDataDirs(source, upgradeID, idl.ClusterType_SOURCE)
	targetDirs := getNonMasterDataDirs(target, upgradeID, idl.ClusterType_TARGET)
	m := make(MarkMap)
	for host, dirs := range sourceDirs {
		m[host] = append(m[host], dirs...)
		m[host] = append(m[host], targetDirs[host]...)
		sort.Strings(m[host]) // makes testing easier
	}

	err := markSegmentDataDirs(agentConns, m)
	if err != nil {
		return xerrors.Errorf("adding marker files: %w", err)
	}

	return nil
}

func markSegmentDataDirs(agentConns []*Connection, markers MarkMap) error {

	wg := sync.WaitGroup{}
	errs := make(chan error, len(agentConns))

	for _, conn := range agentConns {
		conn := conn

		if len(markers[conn.Hostname]) == 0 {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			req := &idl.CreateFilesRequest{Files: markers[conn.Hostname]}
			_, err := conn.AgentClient.CreateFiles(context.Background(), req)
			if err != nil {
				gplog.Error("marking segment data directories on host %s: %s", conn.Hostname, err.Error())
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	var mErr *multierror.Error
	for err := range errs {
		mErr = multierror.Append(mErr, err)
	}

	return mErr.ErrorOrNil()
}

func getNonMasterDataDirs(cluster *greenplum.Cluster, id upgrade.ID, kind idl.ClusterType) MarkMap {
	m := make(MarkMap)

	segs := cluster.SelectSegments(func(seg *greenplum.SegConfig) bool {
		return !seg.IsMaster()
	})

	for _, seg := range segs {
		host := seg.Hostname
		m[host] = append(m[host], upgrade.MarkerFileName(seg.DataDir, id, kind))
	}

	return m
}
