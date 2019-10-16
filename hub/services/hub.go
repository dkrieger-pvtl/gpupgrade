package services

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"
	"github.com/greenplum-db/gpupgrade/hub/upgradestatus/file"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/greenplum-db/gpupgrade/utils/daemon"
	"github.com/greenplum-db/gpupgrade/utils/log"
)

var DialTimeout = 3 * time.Second

// Returned from Hub.Start() if Hub.Stop() has already been called.
var ErrHubStopped = errors.New("hub is stopped")

type Hub struct {
	conf *HubConfig

	agentConns []*Connection
	source     *utils.Cluster
	target     *utils.Cluster
	checklist  upgradestatus.Checklist

	mu     sync.Mutex
	server *grpc.Server
	lis    net.Listener

	// This is used both as a channel to communicate from Start() to
	// Stop() to indicate to Stop() that it can finally terminate
	// and also as a flag to communicate from Stop() to Start() that
	// Stop() had already beed called, so no need to do anything further
	// in Start().
	// Note that when used as a flag, nil value means that Stop() has
	// been called.

	stopped chan struct{}
	daemon  bool
}

type Connection struct {
	Hostname      string
	Conn          *grpc.ClientConn
	CancelContext func()
}

type HubConfig struct {
	CliToHubPort   int
	HubToAgentPort int
	StateDir       string
	LogDir         string
}

func NewHub(sourceCluster *utils.Cluster, targetCluster *utils.Cluster, conf *HubConfig, checklist upgradestatus.Checklist) *Hub {
	h := &Hub{
		stopped:   make(chan struct{}, 1),
		conf:      conf,
		source:    sourceCluster,
		target:    targetCluster,
		checklist: checklist,
	}

	return h
}

// MakeDaemon tells the Hub to disconnect its stdout/stderr streams after
// successfully starting up.
func (h *Hub) MakeDaemon() {
	h.daemon = true
}

func (h *Hub) Start() error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(h.conf.CliToHubPort))
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}

	// Set up an interceptor function to log any panics we get from request
	// handlers.
	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer log.WritePanics()
		return handler(ctx, req)
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor))

	h.mu.Lock()
	if h.stopped == nil {
		// Stop() has already been called; return without serving.
		h.mu.Unlock()
		return ErrHubStopped
	}
	h.server = server
	h.lis = lis
	h.mu.Unlock()

	idl.RegisterCliToHubServer(server, h)
	reflection.Register(server)

	if h.daemon {
		fmt.Printf("Hub started on port %d (pid %d)\n", h.conf.CliToHubPort, os.Getpid())
		daemon.Daemonize()
	}

	err = server.Serve(lis)
	if err != nil {
		err = errors.Wrap(err, "failed to serve")
	}

	// inform Stop() that is it is OK to stop now
	h.stopped <- struct{}{}

	return err
}

func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closeConns()

	if h.server != nil {
		h.server.Stop()
		<-h.stopped // block until it is OK to stop
	}

	// Mark this server stopped so that a concurrent Start() doesn't try to
	// start things up again.
	h.stopped = nil
}

type Dialer func(ctx context.Context, address string) (net.Conn, error)

func (h *Hub) AgentConns(dialer Dialer) ([]*Connection, error) {
	// Lock the mutex to protect against races with Hub.Stop().
	// XXX This is a *ridiculously* broad lock. Have fun waiting for the dial
	// timeout when calling Stop() and AgentConns() at the same time, for
	// instance. We should not lock around a network operation, but it seems
	// like the AgentConns concept is not long for this world anyway.
	h.mu.Lock()
	defer h.mu.Unlock()

	hostnames := h.source.PrimaryHostnames()
	agentConns := make([]*Connection, 0, len(hostnames))

	for _, host := range hostnames {
		address := host + ":" + strconv.Itoa(h.conf.HubToAgentPort)
		ctx, cancelFunc := context.WithTimeout(context.Background(), DialTimeout)
		opts := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithContextDialer(dialer),
			//grpc.FailOnNonTempDialError(true),
		}
		conn, err := grpc.DialContext(ctx, address, opts...)
		if err != nil {
			cancelFunc()
			gplog.Error("failed to dial agent connection to %s: %+v", host, err)
			return nil, xerrors.Errorf("failed to dial agent %w", err)
		}

		agentConns = append(agentConns, &Connection{
			Hostname:      host,
			Conn:          conn,
			CancelContext: cancelFunc,
		})
	}

	return agentConns, nil
}

// Closes all h.agentConns. Callers must hold the Hub's mutex.
func (h *Hub) closeConns() {
	for _, conn := range h.agentConns {
		defer conn.CancelContext()
		currState := conn.Conn.GetState()
		err := conn.Conn.Close()
		if err != nil {
			gplog.Info(fmt.Sprintf("Error closing hub to agent connection. host: %s, err: %s", conn.Hostname, err.Error()))
		}
		conn.Conn.WaitForStateChange(context.Background(), currState)
	}
}

// streamStepWriter extends the standard StepWriter, which only writes state to
// disk, with functionality that sends status updates across the given stream.
// (In practice this stream will be a gRPC CliToHub_XxxServer interface.)
type streamStepWriter struct {
	upgradestatus.StateWriter
	stream messageSender
}

type messageSender interface {
	Send(*idl.Message) error // matches gRPC streaming Send()
}

func sendStatus(stream messageSender, step idl.UpgradeSteps, status idl.StepStatus) {
	// A stream is not guaranteed to remain connected during execution, so
	// errors are explicitly ignored.
	_ = stream.Send(&idl.Message{
		Contents: &idl.Message_Status{&idl.UpgradeStepStatus{
			Step:   step,
			Status: status,
		}},
	})
}

func (s streamStepWriter) MarkInProgress() error {
	if err := s.StateWriter.MarkInProgress(); err != nil {
		return err
	}

	sendStatus(s.stream, s.Code(), idl.StepStatus_RUNNING)
	return nil
}

func (s streamStepWriter) MarkComplete() error {
	if err := s.StateWriter.MarkComplete(); err != nil {
		return err
	}

	sendStatus(s.stream, s.Code(), idl.StepStatus_COMPLETE)
	return nil
}

func (s streamStepWriter) MarkFailed() error {
	if err := s.StateWriter.MarkFailed(); err != nil {
		return err
	}

	sendStatus(s.stream, s.Code(), idl.StepStatus_FAILED)
	return nil
}

// Extracts common hub logic to reset state directory, mark step as in-progress,
// and control status streaming.
func (h *Hub) InitializeStep(step string, stream messageSender) (upgradestatus.StateWriter, error) {
	stepWriter := streamStepWriter{
		h.checklist.GetStepWriter(step),
		stream,
	}

	err := stepWriter.ResetStateDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to reset state directory")
	}

	err = stepWriter.MarkInProgress()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set %s to %s", step, file.InProgress)
	}

	return stepWriter, nil
}
