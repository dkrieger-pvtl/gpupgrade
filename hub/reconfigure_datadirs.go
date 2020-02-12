package hub

import (
	"io"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/step"
)

func (s *Server) ReconfigureDatadirs(stream step.OutStreams) (err error) {
	// Stop the target cluster if it is not already

	// update the datadirs in the cluster
	cmd := execCommand("postgres",
		"--single",
		"-D "+s.Target.MasterDataDir(),
		"-c \"allow_system_table_mods=true\"",
		"postgres",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return xerrors.Errorf("getting stdin pipe for reconfigure datadirs: %w", err)
	}

	errs := make(chan error, 1)
	go func() {
		defer stdin.Close()
		_, err = io.WriteString(stdin, "UPDATE gp_segment_configuration SET datadir = datadir || '_old';")
		errs <- err
	}()

	err = cmd.Run()
	if err != nil {
		errs <- xerrors.Errorf("reconfiguring datadirs: %w", err)
	}

	var multiErr multierror.Error
	for err := range errs {
		multiErr = *multierror.Append(&multiErr, err)
	}

	// rename the target cluster datadirs. _upgrade -> _

	// rename the source cluster datadirs. _ -> _old

	return multiErr.ErrorOrNil()
}
