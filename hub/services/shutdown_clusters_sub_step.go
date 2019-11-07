package services

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/greenplum-db/gpupgrade/idl"

	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

var isPostmasterRunningCmd = exec.Command
var startStopClusterCmd = exec.Command

const (
	SOURCE = "SOURCE"
	TARGET = "TARGET"
)

func (h *Hub) ShutdownClusters(stream messageSender, log io.Writer, inputs ...string) error {
	var shutdownErr error

	switch inputs[0] {
	case SOURCE:
		err := StopCluster(stream, log, h.source)
		if err != nil {
			shutdownErr = multierror.Append(shutdownErr, errors.Wrap(err, "failed to stop source cluster"))
		}
	case TARGET:
		err := StopCluster(stream, log, h.target)
		if err != nil {
			shutdownErr = multierror.Append(shutdownErr, errors.Wrap(err, "failed to stop target cluster"))
		}
	}

	return shutdownErr
}

func StopCluster(stream messageSender, log io.Writer, cluster *utils.Cluster) error {
	return startStopCluster(stream, log, cluster, true)
}
func StartCluster(stream messageSender, log io.Writer, cluster *utils.Cluster) error {
	return startStopCluster(stream, log, cluster, false)
}

func startStopCluster(stream messageSender, log io.Writer, cluster *utils.Cluster, stop bool) error {
	// TODO: why can't we call IsPostmasterRunning for the !stop case?  If we do, we get this on the pipeline:
	// Usage: pgrep [-flvx] [-d DELIM] [-n|-o] [-P PPIDLIST] [-g PGRPLIST] [-s SIDLIST]
	// [-u EUIDLIST] [-U UIDLIST] [-G GIDLIST] [-t TERMLIST] [PATTERN]
	//  pgrep: pidfile not valid
	// TODO: should we actually return an error if we try to gpstop an already stopped cluster?
	cmdName := "gpstart"
	if stop {
		cmdName = "gpstop"
		err := IsPostmasterRunning(stream, log, cluster)
		if err != nil {
			return err
		}
	}

	cmd := startStopClusterCmd("bash", "-c",
		fmt.Sprintf("source %[1]s/../greenplum_path.sh && %[1]s/%[2]s -a -d %[3]s",
			cluster.BinDir,
			cmdName,
			cluster.MasterDataDir(),
		))

	mux := newMultiplexedStream(stream, log)
	cmd.Stdout = mux.NewStreamWriter(idl.Chunk_STDOUT)
	cmd.Stderr = mux.NewStreamWriter(idl.Chunk_STDERR)

	return cmd.Run()
}

func IsPostmasterRunning(stream messageSender, log io.Writer, cluster *utils.Cluster) error {
	cmd := isPostmasterRunningCmd("bash", "-c",
		fmt.Sprintf("pgrep -F %s/postmaster.pid",
			cluster.MasterDataDir(),
		))

	mux := newMultiplexedStream(stream, log)
	cmd.Stdout = mux.NewStreamWriter(idl.Chunk_STDOUT)
	cmd.Stderr = mux.NewStreamWriter(idl.Chunk_STDERR)

	return cmd.Run()
}