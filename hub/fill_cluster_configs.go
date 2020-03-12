package hub

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	"github.com/greenplum-db/gpupgrade/db"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
)

// create source/target clusters, write to disk and re-read from disk to make sure it is "durable"
func (s *Server) FillClusterConfigsSubStep(config *Config, conn *sql.DB, _ step.OutStreams, request *idl.InitializeRequest, saveConfig func() error) error {
	if err := CheckSourceClusterConfiguration(conn); err != nil {
		return err
	}

	// XXX ugly; we should just use the conn we're passed, but our DbConn
	// concept (which isn't really used) gets in the way
	dbconn := db.NewDBConn("localhost", int(request.SourcePort), "template1")
	source, err := utils.ClusterFromDB(dbconn, request.SourceBinDir)
	if err != nil {
		return errors.Wrap(err, "could not retrieve source configuration")
	}

	config.Source = source
	config.Target = &utils.Cluster{BinDir: request.TargetBinDir}
	config.UseLinkMode = request.UseLinkMode

	var ports []int
	for _, p := range request.Ports {
		ports = append(ports, int(p))
	}

	s.TargetInitializeConfig, err = AssignDatadirsAndPorts(s.Source, ports)
	if err != nil {
		return err
	}

	if err := saveConfig(); err != nil {
		return err
	}

	return nil
}

func AssignDatadirsAndPorts(source *utils.Cluster, ports []int) (InitializeConfig, error) {
	if len(ports) == 0 {
		port := 50432
		numberofPrimaries := len(source.Primaries) // NOTE: source.Primaries includes the master
		if (numberofPrimaries + port) > 65535 {
			numberofPrimaries = 65535 - port
		}

		for i := 0; i < numberofPrimaries; i++ {
			ports = append(ports, port)
			port++
		}
	} else {
		ports = sanitize(ports)
	}

	initializeConfig, err := assignPrimaryDatadirsAndCustomPorts(source, ports)
	if err != nil {
		return initializeConfig, err
	}

	// We copy the source standby and mirror segment configs exactly, since
	// we initialize the mirrors and standby at the end of the finalize step now.
	for _, content := range source.ContentIDs {
		seg, ok := source.Mirrors[content]
		if ok && content == -1 {
			initializeConfig.Standby = seg
		} else if ok {
			initializeConfig.Mirrors = append(initializeConfig.Mirrors, seg)
		}
	}

	return initializeConfig, err
}

// can return an error if we run out of ports to use
func assignPrimaryDatadirsAndCustomPorts(source *utils.Cluster, ports []int) (InitializeConfig, error) {
	targetInitializeConfig := InitializeConfig{}

	nextPortIndex := 0

	if master, ok := source.Primaries[-1]; ok {
		// Reserve a port for the master.
		if nextPortIndex > len(ports)-1 {
			return InitializeConfig{}, errors.New("not enough ports")
		}
		master.Port = ports[nextPortIndex]
		master.DataDir = upgradeDataDir(master.DataDir)
		targetInitializeConfig.Master = master
		nextPortIndex++
	}

	portIndexByHost := make(map[string]int)

	for _, content := range source.ContentIDs {
		// Skip the master segment
		if content == -1 {
			continue
		}

		segment := source.Primaries[content]

		if portIndex, ok := portIndexByHost[segment.Hostname]; ok {
			if portIndex > len(ports)-1 {
				return InitializeConfig{}, errors.New("not enough ports")
			}
			segment.Port = ports[portIndex]
			portIndexByHost[segment.Hostname]++
		} else {
			if nextPortIndex > len(ports)-1 {
				return InitializeConfig{}, errors.New("not enough ports")
			}
			segment.Port = ports[nextPortIndex]
			portIndexByHost[segment.Hostname] = nextPortIndex + 1
		}
		segment.DataDir = upgradeDataDir(segment.DataDir)

		targetInitializeConfig.Primaries = append(targetInitializeConfig.Primaries, segment)
	}

	return targetInitializeConfig, nil
}

// sanitize sorts and deduplicates a slice of port numbers.
func sanitize(ports []int) []int {
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })

	dedupe := ports[:0] // point at the same backing array

	var last int
	for i, port := range ports {
		if i == 0 || port != last {
			dedupe = append(dedupe, port)
		}
		last = port
	}

	return dedupe
}

func getAgentPath() (string, error) {
	hubPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Dir(hubPath), "gpupgrade"), nil
}
