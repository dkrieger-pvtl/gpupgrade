package hub

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/pkg/errors"

	"github.com/greenplum-db/gpupgrade/idl"
)

type AgentBroker interface {
	ReconfigureDataDirectories(hostname string, renamePairs []*idl.RenamePair) error
}

type AgentBrokerGRPC struct {
	agentConnections []*Connection
	context          context.Context
}

func (broker *AgentBrokerGRPC) ReconfigureDataDirectories(hostname string, renamePairs []*idl.RenamePair) error {
	var connection *Connection

	// find the client for this agent's host
	for _, c := range broker.agentConnections {
		if c.Hostname == hostname {
			connection = c
			break
		}
	}

	if connection == nil {
		return errors.New(fmt.Sprintf("No agent connections for hostname=%v", hostname))
	}

	_, err := connection.AgentClient.ReconfigureDataDirectories(
		broker.context,
		&idl.ReconfigureDataDirRequest{
			Pairs: renamePairs,
		},
	)

	return err
}
