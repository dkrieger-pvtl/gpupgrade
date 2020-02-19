package hub

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/greenplum-db/gpupgrade/idl"
)

type AgentBroker interface {
	ReconfigureDataDirectories(hostname string, renamePairs []*idl.RenamePair) error
}

type AgentBrokerGRPC struct {
	agentConnections []*Connection
}

func (s *AgentBrokerGRPC) ReconfigureDataDirectories(hostname string, renamePairs []*idl.RenamePair) error {
	var connection *Connection

	// find the client for this agent's host
	for _, c := range s.agentConnections {
		if c.Hostname == hostname {
			connection = c
			break
		}
	}

	if connection == nil {
		return errors.New(fmt.Sprintf("No agent connections for hostname=%v", hostname))
	}

	_, err := connection.AgentClient.ReconfigureDataDirectories(
		context.Background(),
		&idl.ReconfigureDataDirRequest{
			Pair: renamePairs,
		},
		nil,
	)

	return err
}
