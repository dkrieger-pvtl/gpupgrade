package services

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/hub/upgradestatus"
	"github.com/greenplum-db/gpupgrade/idl"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

const (
	SedAndMvString = "sed 's/port=%d/port=%d/' %[3]s/postgresql.conf > %[3]s/postgresql.conf.updated && " +
		"mv %[3]s/postgresql.conf %[3]s/postgresql.conf.bak && " +
		"mv %[3]s/postgresql.conf.updated %[3]s/postgresql.conf"
)

func (h *Hub) UpgradeReconfigurePorts(ctx context.Context, in *idl.UpgradeReconfigurePortsRequest) (*idl.UpgradeReconfigurePortsReply, error) {
	gplog.Info("starting %s", upgradestatus.RECONFIGURE_PORTS)

	step, err := h.InitializeStep(upgradestatus.RECONFIGURE_PORTS)
	if err != nil {
		gplog.Error(err.Error())
		return &idl.UpgradeReconfigurePortsReply{}, err
	}

	if err := h.reconfigurePorts(); err != nil {
		gplog.Error(err.Error())
		step.MarkFailed()
		return &idl.UpgradeReconfigurePortsReply{}, err
	}

	step.MarkComplete()
	return &idl.UpgradeReconfigurePortsReply{}, nil
}

// reconfigurePorts executes the tricky sequence of operations required to
// change the ports on a cluster:
//    1). bring down the cluster
//    2). bring up the master(fts will not "freak out", etc)
//    3). rewrite gp_segment_configuration with the updated port number
//    4). modify the master's config file to use the new port
//    5). bring down the master
//    6). bring up the cluster
func (h *Hub) reconfigurePorts() (err error) {
	sedCommand := fmt.Sprintf(SedAndMvString, h.target.MasterPort(), h.source.MasterPort(), h.target.MasterDataDir())
	gplog.Debug("executing command: %+v", sedCommand) // TODO: Move this debug log into ExecuteLocalCommand()

	// 1). bring down the cluster
	err = StopCluster(h.target)
	if err != nil {
		return xerrors.Errorf("%s failed to stop cluster: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}

	// 2). bring up the master(fts will not "freak out", etc)
	startScript := fmt.Sprintf("source %s/../greenplum_path.sh && %s/gpstart -am -d %s",
		h.target.BinDir, h.target.BinDir, h.target.MasterDataDir())
	cmd := exec.Command("bash", "-c", startScript)
	err = cmd.Run()
	if err != nil {
		return xerrors.Errorf("%s failed to start target cluster in utility mode: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}

	// 3). rewrite gp_segment_configuration with the updated port number
	connURI := fmt.Sprintf("postgresql://localhost:%d/template1?gp_session_role=utility&search_path=", h.target.MasterPort())
	targetDB, err := sql.Open("pgx", connURI)
	defer func() {
		targetDB.Close() //TODO: return multierror here to capture err from Close()
	}()
	if err != nil {
		return xerrors.Errorf("%s failed to open connection to utility master: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}
	err = ClonePortsFromCluster(targetDB, h.source.Cluster)
	if err != nil {
		return xerrors.Errorf("%s failed to clone ports: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}

	// 4). rewrite the "port" field in the master's postgresql.conf
	_, err = targetDB.Exec("ALTER SYSTEM SET port TO ?", h.source.MasterPort())
	if err != nil {
		return xerrors.Errorf("%s failed to clone ports: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}

	// 5). bring down the master and bring back up the cluster
	// # TODO: we apparently cannot `gpstop -air` from master-only to full cluster
	// #   run these as separate commands
	stopScript := fmt.Sprintf("source %s/../greenplum_path.sh && %s/gpstop -air -d %s",
		h.target.BinDir, h.target.BinDir, h.target.MasterDataDir())
	cmd = exec.Command("bash", "-c", stopScript)
	err = cmd.Run()
	if err != nil {
		return xerrors.Errorf("%s failed to stop target cluster in utility mode: %w",
			upgradestatus.RECONFIGURE_PORTS, err)
	}

	return nil
}
