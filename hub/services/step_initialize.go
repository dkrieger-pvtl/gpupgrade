package services

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"

	"google.golang.org/grpc"

	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/db"
	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (h *Hub) Initialize(ctx context.Context, in *idl.InitializeRequest) (*idl.InitializeReply, error) {
	err := h.fillClusterConfigsSubStep(in.OldBinDir, in.NewBinDir, int(in.OldPort))
	if err != nil {
		return &idl.InitializeReply{}, err
	}

	err = h.startAgentsSubStep()
	if err != nil {
		return &idl.InitializeReply{}, err
	}

	return &idl.InitializeReply{}, nil
}

// create old/new clusters, write to disk and re-read from disk to make sure it is "durable"
func (h *Hub) fillClusterConfigsSubStep(oldBinDir, newBinDir string, oldPort int) error {
	gplog.Info("starting %s", upgradestatus.CONFIG)

	step, err := h.InitializeStep(upgradestatus.CONFIG)
	if err != nil {
		gplog.Error(err.Error())
		return err
	}

	err = h.fillClusterConfigs(oldBinDir, newBinDir, oldPort)

	if err != nil {
		gplog.Error(err.Error())
		step.MarkFailed()
		return err
	}

	step.MarkComplete()
	return nil
}

func (h *Hub) startAgentsSubStep() error {
	gplog.Info("starting %s", upgradestatus.START_AGENTS)

	step, err := h.InitializeStep(upgradestatus.START_AGENTS)
	if err != nil {
		gplog.Error(err.Error())
		return err
	}

	err = StartAgents(h.source, h.target, h.conf.StateDir)
	if err != nil {
		gplog.Error(err.Error())
		step.MarkFailed()
		return err
	}

	step.MarkComplete()
	return nil
}

func (h *Hub) fillClusterConfigs(oldBinDir, newBinDir string, oldPort int) error {
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

func StartAgents(source *utils.Cluster, target *utils.Cluster, stateDir string) error {
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

func (h *Hub) RestartAgents(ctx context.Context, in *idl.RestartAgentsRequest) (*idl.RestartAgentsReply, error) {
	var wg sync.WaitGroup
	hostnames := h.source.GetHostnames()
	restartedHosts := make(chan string, len(hostnames))
	errs := make(chan error, len(hostnames))

	for _, host := range hostnames {
		wg.Add(1)
		go func() {
			defer wg.Done()

			address := host + ":" + strconv.Itoa(h.conf.HubToAgentPort)
			timeoutCtx, cancelFunc := context.WithTimeout(ctx, 3*time.Second)
			opts := []grpc.DialOption{
				grpc.WithBlock(),
				grpc.WithInsecure(),
				grpc.FailOnNonTempDialError(true),
			}
			conn, err := grpc.DialContext(timeoutCtx, address, opts...)
			cancelFunc()
			if err == nil {
				err = conn.Close()
				if err != nil {
					gplog.Error("when checking agent for connectivity to %s, failed to close: %+v", host, err)
				}

				return
			}

			gplog.Debug("when checking agent for connectivity to %s, could not dial: %+v", host, err)
			gplog.Info("starting agent on %s", host)

			agentPath, err := getAgentPath()
			if err != nil {
				errs <- err
				return
			}
			cmd := execCommand("ssh", host,
				fmt.Sprintf("bash -c \"%s --daemonize --state-directory %s\"", agentPath, h.conf.StateDir))
			stdout, err := cmd.Output()
			if err != nil {
				errs <- err
				return
			}

			gplog.Debug(string(stdout))
			restartedHosts <- host
		}()
	}

	wg.Wait()
	close(errs)
	close(restartedHosts)

	var multiErr *multierror.Error
	for err := range errs {
		multiErr = multierror.Append(multiErr, err)
	}

	var reply idl.RestartAgentsReply
	for h := range restartedHosts {
		reply.AgentHosts = append(reply.AgentHosts, h)
	}

	return &reply, multiErr.ErrorOrNil()
}
