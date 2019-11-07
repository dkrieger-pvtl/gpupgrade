package services

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/db"
	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/pkg/errors"
)

type InitializeStream struct {
	stream messageSender
	log    *os.File
}

func (h *Hub) Initialize(in *idl.InitializeRequest, stream idl.CliToHub_InitializeServer) error {
	log, err := utils.System.OpenFile(
		filepath.Join(utils.GetStateDir(), "initialize.log"),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := log.Close(); closeErr != nil {
			err = multierror.Append(err,
				xerrors.Errorf("failed to close initialize log: %w", closeErr))
		}
	}()

	initializeStream := &InitializeStream{stream: stream, log: log}

	_, err = log.WriteString("\nInitialize in progress.\n")
	if err != nil {
		return xerrors.Errorf("failed writing to initialize log: %w", err)
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.CONFIG, h.fillClusterConfigsSubStep,
		in.OldBinDir, in.NewBinDir, strconv.Itoa(int(in.OldPort)))
	if err != nil {
		return err
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.START_AGENTS, h.startAgentsSubStep)
	if err != nil {
		return err
	}

	return nil
}

func (h *Hub) InitializeCreateCluster(in *idl.InitializeCreateClusterRequest, stream idl.CliToHub_InitializeCreateClusterServer) error {
	log, err := utils.System.OpenFile(
		filepath.Join(utils.GetStateDir(), "initialize.log"),
		os.O_WRONLY|os.O_APPEND,
		0600,
	)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := log.Close(); closeErr != nil {
			err = multierror.Append(err,
				xerrors.Errorf("failed to close initialize log: %w", closeErr))
		}
	}()

	initializeStream := &InitializeStream{stream: stream, log: log}

	_, err = log.WriteString("\nInitialize hub in progress.\n")
	if err != nil {
		return xerrors.Errorf("failed writing to initialize log for hub: %w", err)
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.CREATE_TARGET_CONFIG,
		func(_ messageSender, _ io.Writer, _ ...string) error {
			return h.GenerateInitsystemConfig()
		})
	if err != nil {
		return err
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.SHUTDOWN_SOURCE_CLUSTER,
		func(stream messageSender, log io.Writer, _ ...string) error {
			return StopCluster(stream, log, h.source)
		})
	if err != nil {
		return err
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.INIT_TARGET_CLUSTER, h.CreateTargetCluster)
	if err != nil {
		return err
	}

	err = h.InitializeSubStep(initializeStream, upgradestatus.SHUTDOWN_TARGET_CLUSTER, h.ShutdownClusters, TARGET)
	if err != nil {
		return err
	}

	// pg_upgrade --check both
	err = h.InitializeSubStep(initializeStream, upgradestatus.CHECK_MASTER, h.UpgradeMaster, MASTER_CHECK)
	if err != nil {
		return err
	}
	err = h.ConvertPrimaries(true)
	if err != nil {
		return err
	}

	return nil
}

// TODO: is there a way to pass in a variadic arg but still get type checking?
func (h *Hub) InitializeSubStep(initializeStream *InitializeStream, subStep string,
	subStepFunc func(stream messageSender, log io.Writer, inputs ...string) error,
	args ...string) error {

	gplog.Info("starting %s", subStep)
	_, err := initializeStream.log.WriteString(fmt.Sprintf("\nStarting %s...\n\n", subStep))
	if err != nil {
		return xerrors.Errorf("failed writing to initialize log: %w", err)
	}

	step, err := h.InitializeStep(subStep, initializeStream.stream)
	if err != nil {
		gplog.Error(err.Error())
		return err
	}

	err = subStepFunc(initializeStream.stream, initializeStream.log, args...)
	if err != nil {
		gplog.Error(err.Error())
		step.MarkFailed()
	} else {
		step.MarkComplete()
	}

	return nil
}

// create old/new clusters, write to disk and re-read from disk to make sure it is "durable"
func (h *Hub) fillClusterConfigsSubStep(stream messageSender, log io.Writer, args ...string) error {
	oldBinDir := args[0]
	newBinDir := args[1]
	oldPort, _ := strconv.Atoi(args[2])

	source := &utils.Cluster{BinDir: path.Clean(oldBinDir), ConfigPath: filepath.Join(h.conf.StateDir, utils.SOURCE_CONFIG_FILENAME)}
	dbConn := db.NewDBConn("localhost", oldPort, "template1")
	defer dbConn.Close()
	err := ReloadAndCommitCluster(source, dbConn)
	if err != nil {
		return err
	}

	emptyCluster := cluster.NewCluster([]cluster.SegConfig{})
	target := &utils.Cluster{Cluster: emptyCluster, BinDir: path.Clean(newBinDir), ConfigPath: filepath.Join(h.conf.StateDir, utils.TARGET_CONFIG_FILENAME)}
	err = target.Commit()
	if err != nil {
		return errors.Wrap(err, "Unable to save target cluster configuration")
	}

	// XXX: This is really not necessary as we are just verifying that
	// the configuration that we just wrote is readable.
	errSource := source.Load()
	errTarget := target.Load()
	if errSource != nil && errTarget != nil {
		errBoth := errors.Errorf("Source error: %s\nTarget error: %s", errSource.Error(), errTarget.Error())
		return errors.Wrap(errBoth, "Unable to load source or target cluster configuration")
	} else if errSource != nil {
		return errors.Wrap(errSource, "Unable to load source cluster configuration")
	} else if errTarget != nil {
		return errors.Wrap(errTarget, "Unable to load target cluster configuration")
	}

	// link in source/target to hub
	h.source = source
	h.target = target

	return err
}

// ReloadAndCommitCluster() will fill in a utils.Cluster using a database
// connection and additionally write the results to disk.
func ReloadAndCommitCluster(cluster *utils.Cluster, conn *dbconn.DBConn) error {
	newCluster, err := utils.ClusterFromDB(conn, cluster.BinDir, cluster.ConfigPath)
	if err != nil {
		return errors.Wrap(err, "could not retrieve cluster configuration")
	}

	*cluster = *newCluster
	err = cluster.Commit()
	if err != nil {
		return errors.Wrap(err, "could not save cluster configuration")
	}

	return nil
}

func getAgentPath() (string, error) {
	hubPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(hubPath), "gpupgrade_agent"), nil
}

// TODO: use the implementation in RestartAgents() for this function and combine them
func (h *Hub) startAgentsSubStep(stream messageSender, log io.Writer, inputs ...string) error {
	source := h.source
	stateDir := h.conf.StateDir

	// XXX If there are failures, does it matter what agents have successfully
	// started, or do we just want to stop all of them and kick back to the
	// user?
	logStr := "start agents on master and hosts"

	agentPath, err := getAgentPath()
	if err != nil {
		return errors.Errorf("failed to get the hub executable path %v", err)
	}

	// XXX State directory handling on agents needs to be improved. See issue
	// #127: all agents will silently recreate that directory if it doesn't
	// already exist. Plus, ExecuteOnAllHosts() doesn't let us control whether
	// we execute locally or via SSH for the master, so we don't know whether
	// GPUPGRADE_HOME is going to be inherited.
	runAgentCmd := func(contentID int) string {
		return agentPath + " --daemonize --state-directory " + stateDir
	}

	errStr := "Failed to start all gpupgrade_agents"

	remoteOutput, err := source.ExecuteOnAllHosts(logStr, runAgentCmd)
	if err != nil {
		return errors.Wrap(err, errStr)
	}

	errMessage := func(contentID int) string {
		return fmt.Sprintf("Could not start gpupgrade_agent on segment with contentID %d", contentID)
	}
	source.CheckClusterError(remoteOutput, errStr, errMessage, true)

	// Agents print their port and PID to stdout; log them for posterity.
	for content, output := range remoteOutput.Stdouts {
		if remoteOutput.Errors[content] == nil {
			gplog.Info("[%s] %s", source.Segments[content].Hostname, output)
		}
	}

	if remoteOutput.NumErrors > 0 {
		// CheckClusterError() will have already logged each error.
		return errors.New("could not start agents on segment hosts; see log for details")
	}

	return nil
}
