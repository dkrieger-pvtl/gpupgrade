package agentclient

import (
	"sync"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

func TestStartAgent(suite *testing.T) {
	suite.Run("", func(test *testing.T) {
		gplog.InitializeLogging(
			"gpupgrade hub",
			"/tmp/foobar",
		)

		wg := &sync.WaitGroup{}
		errorChannel := make(chan error, 1)

		client := New(wg, errorChannel)

		client.StartAgent("foobar", "somestatedir")
	})
}
