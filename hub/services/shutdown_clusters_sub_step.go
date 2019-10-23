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
	err := IsPostmasterRunning(stream, log, cluster)
	if stop {
		if err != nil {
			return err
		}
	} else {
		if err == nil {
			return errors.New("cluster already up")
		}
	}

	cmdName := "gpstart"
	if stop {
		cmdName = "gpstop"
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
