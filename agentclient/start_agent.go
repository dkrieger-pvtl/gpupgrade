package agentclient

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"google.golang.org/grpc"
)

type agentStarter struct {
	wg           *sync.WaitGroup
	errorChannel chan error
}

type dialerFunc func(context.Context, string) (net.Conn, error)

var execCommand = exec.Command

func getAgentPath() (string, error) {
	hubPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(hubPath), "gpupgrade"), nil
}

func (a *agentStarter) isAgentRunning(host string) bool {
	port := 6416
	var dialer dialerFunc
	ctx := context.Background()

	// Is agent running
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

		return true
	}

	gplog.Debug("failed to dial agent on %s: %+v", host, err)
	gplog.Info("starting agent on %s", host)

	return false
}

func startAgent(hostname, stateDir string) error {
	agentPath, err := getAgentPath()

	if err != nil {
		return err
	}

	cmd := execCommand("ssh", hostname,
		fmt.Sprintf("bash -c \"%s agent --daemonize --state-directory %s\"",
			agentPath,
			stateDir))

	stdout, err := cmd.Output()

	gplog.Debug(string(stdout))

	return nil
}

func (a *agentStarter) StartAgent(hostname, stateDir string) {
	fmt.Printf("starting agent on host %s using stateDir %s", hostname, stateDir)

	a.wg.Add(1)

	go func() {
		a.wg.Done()

		if !a.isAgentRunning(hostname) {
			startAgentError := startAgent(hostname, stateDir)

			if startAgentError != nil {
				a.errorChannel <- startAgentError
				return
			}
		}
	}()

}

func New(wg *sync.WaitGroup, errorChannel chan error) *agentStarter {
	return &agentStarter{
		wg,
		errorChannel,
	}
}
