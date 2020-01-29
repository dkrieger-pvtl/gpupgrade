package hub

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/greenplum-db/gpupgrade/db"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
)

type AgentStarter interface {
	StartAgent(hostname, stateDir string) (err error)
}
type agentStarter struct{}

func (h *Hub) Initialize(in *idl.InitializeRequest, stream idl.CliToHub_InitializeServer) (err error) {
	s, err := BeginStep(h.StateDir, "initialize", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := s.Finish(); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("initialize: %s", err))
		}
	}()

	s.Run(idl.Substep_CONFIG, func(stream step.OutStreams) error {
		return h.fillClusterConfigsSubStep(stream, in)
	})

	s.Run(idl.Substep_START_AGENTS, func(stream step.OutStreams) error {
		return StartAgentsSubStep(h.Source.GetHostnames(), h.StateDir, &agentStarter{})
	})

	return s.Err()
}

func (h *Hub) InitializeCreateCluster(in *idl.InitializeCreateClusterRequest, stream idl.CliToHub_InitializeCreateClusterServer) (err error) {
	s, err := BeginStep(h.StateDir, "initialize", stream)
	if err != nil {
		return err
	}

	defer func() {
		if ferr := s.Finish(); ferr != nil {
			err = multierror.Append(err, ferr).ErrorOrNil()
		}

		if err != nil {
			gplog.Error(fmt.Sprintf("initialize: %s", err))
		}
	}()

	var targetMasterPort int
	s.Run(idl.Substep_CREATE_TARGET_CONFIG, func(_ step.OutStreams) error {
		var err error
		targetMasterPort, err = h.GenerateInitsystemConfig(in.Ports)
		return err
	})

	s.Run(idl.Substep_SHUTDOWN_SOURCE_CLUSTER, func(stream step.OutStreams) error {
		return StopCluster(stream, h.Source)
	})

	s.Run(idl.Substep_INIT_TARGET_CLUSTER, func(stream step.OutStreams) error {
		return h.CreateTargetCluster(stream, targetMasterPort)
	})

	s.Run(idl.Substep_SHUTDOWN_TARGET_CLUSTER, func(stream step.OutStreams) error {
		return h.ShutdownCluster(stream, false)
	})

	s.Run(idl.Substep_CHECK_UPGRADE, func(stream step.OutStreams) error {
		return h.CheckUpgrade(stream)
	})

	return s.Err()
}

// create old/new clusters, write to disk and re-read from disk to make sure it is "durable"
func (h *Hub) fillClusterConfigsSubStep(_ step.OutStreams, request *idl.InitializeRequest) error {
	conn := db.NewDBConn("localhost", int(request.OldPort), "template1")
	defer conn.Close()

	var err error
	h.Source, err = utils.ClusterFromDB(conn, request.OldBinDir)
	if err != nil {
		return errors.Wrap(err, "could not retrieve source configuration")
	}

	h.Target = &utils.Cluster{Cluster: new(cluster.Cluster), BinDir: request.NewBinDir}
	h.UseLinkMode = request.UseLinkMode

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func getAgentPath() (string, error) {
	hubPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(hubPath), "gpupgrade"), nil
}

// TODO: use the implementation in RestartAgents() for this function and combine them
func StartAgentsSubStep(hostnames []string, stateDir string, agentStarter AgentStarter) (err error) {

	for _, host := range hostnames {
		nErr := agentStarter.StartAgent(host, stateDir)
		if nErr != nil {
			err = multierror.Append(err, nErr).ErrorOrNil()
		}
	}
	return err
}

func (a *agentStarter) StartAgent(hostname, stateDir string) (err error) {
	fmt.Printf("starting agent on host %s using stateDir %s", hostname, stateDir)
	err = RestartAgents2(context.Background(), nil, hostname, 6416, stateDir)
	return err
}

func RestartAgents2(ctx context.Context,
	dialer func(context.Context, string) (net.Conn, error),
	hostname string,
	port int,
	stateDir string) error {

	var wg sync.WaitGroup
	errs := make(chan error, 1)

	wg.Add(1)
	go func(host string) {
		defer wg.Done()

		address := host + ":" + strconv.Itoa(port)
		timeoutCtx, cancelFunc := context.WithTimeout(ctx, 3*time.Second)
		opts := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.FailOnNonTempDialError(true),
		}
		if dialer != nil {
			opts = append(opts, grpc.WithContextDialer(dialer))
		}
		conn, err := grpc.DialContext(timeoutCtx, address, opts...)
		cancelFunc()
		if err == nil {
			err = conn.Close()
			if err != nil {
				gplog.Error("failed to close agent connection to %s: %+v", host, err)
			}
			return
		}

		gplog.Debug("failed to dial agent on %s: %+v", host, err)
		gplog.Info("starting agent on %s", host)

		agentPath, err := getAgentPath()
		if err != nil {
			errs <- err
			return
		}
		cmd := execCommand("ssh", host,
			fmt.Sprintf("bash -c \"%s agent --daemonize --state-directory %s\"", agentPath, stateDir))
		stdout, err := cmd.Output()
		if err != nil {
			errs <- err
			return
		}

		gplog.Debug(string(stdout))
	}(hostname)

	wg.Wait()
	close(errs)

	var multiErr *multierror.Error
	for err := range errs {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr.ErrorOrNil()
}
