package hub

import (
	"context"
	"io"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
)

func (s *Server) ReconfigureDatadirs(stream step.OutStreams) (err error) {
	// Stop the target cluster if it is not already

	// update the datadirs in the cluster
	cmd := execCommand("postgres",
		"--single",
		"-D "+s.Target.MasterDataDir(),
		"-c \"allow_system_table_mods=true\"",
		"postgres",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return xerrors.Errorf("getting stdin pipe for reconfigure datadirs: %w", err)
	}

	errs := make(chan error, 2)
	go func() {
		defer stdin.Close()
		_, err = io.WriteString(stdin, "UPDATE gp_segment_configuration SET datadir = datadir || '_old';")
		errs <- err
	}()

	err = cmd.Run()
	if err != nil {
		errs <- xerrors.Errorf("reconfiguring datadirs: %w", err)
	}

	for cErr := range errs {
		err = multierror.Append(err, cErr).ErrorOrNil()
	}

	//noop := func(path string) string {
	//	return path
	//}

	//// rename the target cluster datadirs. _upgrade -> _
	//err = renameDataDirectories(s.agentConns, s.Target, upgradeDataDir, noop)
	//if err != nil {
	//	return xerrors.Errorf("renaming target directories: %w")
	//}
	//
	//// rename the source cluster datadirs. _ -> _old
	//err = renameDataDirectories(s.agentConns, s.Source, noop, oldDataDir)
	//if err != nil {
	//	return xerrors.Errorf("renaming source directories: %w")
	//}

	return err
}

func renameDataDirectories(agentConns []*Connection,
	cluster *utils.Cluster,
	originDataDirFunc func(path string) string,
	destDataDirFunc func(path string) string) error {

	wg := sync.WaitGroup{}
	errs := make(chan error, len(agentConns))

	for _, conn := range agentConns {
		wg.Add(1)

		go func(c *Connection) {
			defer wg.Done()

			segments, err := cluster.SegmentsOn(c.Hostname)
			if err != nil {
				errs <- err
				return
			}

			req := new(idl.ReconfigureDataDirRequest)
			for _, seg := range segments {
				pair := idl.Pair{
					OriginDataDir:      filepath.Dir(originDataDirFunc(seg.DataDir)),
					DestinationDataDir: filepath.Dir(destDataDirFunc(seg.DataDir))}

				req.Pair = append(req.Pair, &pair)
			}

			_, err = c.AgentClient.ReconfigureDataDirectories(context.Background(), req)
			if err != nil {
				gplog.Error("Error creating segment data directories on host %s: %s",
					c.Hostname, err.Error())
				errs <- err
			}
		}(conn)
	}

	wg.Wait()
	close(errs)

	var err error

	for e := range errs {
		err = multierror.Append(err, e)
	}

	return err
}
