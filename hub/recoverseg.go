package hub

import (
	"fmt"
	"os/exec"

	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
)

var RecoversegCmd = exec.Command

func Recoverseg(stream step.OutStreams, cluster *utils.Cluster) error {
	// TODO: consider running with `-B 16` or a cluster-size scaled number
	cmd := RecoversegCmd("bash", "-c",
		fmt.Sprintf("source %[1]s/../greenplum_path.sh && %[1]s/gprecoverseg -a",
			cluster.BinDir,
		))

	cmd.Stdout = stream.Stdout()
	cmd.Stderr = stream.Stderr()

	return cmd.Run()
}
