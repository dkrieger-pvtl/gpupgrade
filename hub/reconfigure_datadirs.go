package hub

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
)

type SegmentConfigModification struct {
	newDataDir string
	dbid       int
}

func ModifySegmentCatalog(cluster *utils.Cluster, modifications []utils.SegConfig) error {
	if len(modifications) == 0 {
		return nil
	}

	return WithinDbTransaction(cluster, func(transaction *sql.Tx) error {
		for _, mod := range modifications {
			_, _ = transaction.Exec("update gp_segment_configuration set datadir = $1 where dbid = $2",
				mod.DataDir, mod.DbID)
		}

		return nil
	})
}

func ChangeDataDirsInCatalog(c *utils.Cluster,
	transformFunc func(text string) string,
	modifySegmentCatalog func(*utils.Cluster, []utils.SegConfig) error) error {

	var mods []utils.SegConfig

	for _, segConfig := range c.Primaries {
		segConfig := segConfig
		segConfig.DataDir = transformFunc(segConfig.DataDir)
		mods = append(mods, segConfig)
	}

	return modifySegmentCatalog(c, mods)
}

//func renameDataDirsInDatabase(cluster *utils.Cluster) error {
//	mods := "-c \"allow_system_table_mods=true\""
//	if cluster.Version.SemVer.Major < 6 {
//		mods = "-c \"allow_system_table_mods='DML'\""
//	}
//
//	cmd := execCommand("postgres",
//		"--single",
//		"-D "+cluster.MasterDataDir(),
//		mods,
//		"template1",
//	)
//	stdin, err := cmd.StdinPipe()
//	if err != nil {
//		return xerrors.Errorf("stdin pipe for renaming data directories: %w", err)
//	}
//
//	// TODO: Fix SQL to append _old to the correct directory.
//	//Normal to _upgrade:
//	//UPDATE table SET field = rtrim(substring(field FROM '^.+/'), '/') || '_upgrade/' || regexp_replace(field, '^.+/', '');
//	//Normal to _old:
//	//UPDATE table SET field = rtrim(substring(field FROM '^.+/'), '/') || '_old/' || regexp_replace(field, '^.+/', '');
//	//_upgrade to Normal:
//	//UPDATE table SET field = rtrim(substring(field FROM '^.+/'), '_upgrade/') || '/' || regexp_replace(field, '^.+/', '');
//	//_upgrade to _old:
//	//UPDATE table SET field = rtrim(substring(field FROM '^.+/'), '_upgrade/') || '_old/' || regexp_replace(field, '^.+/', '');
//	//...where "table" is "gp_segment_configuration" in 6 and "pg_filespace_entry" in 5, and "field" is "datadir" in 6 and "fselocation" in 5.
//	// You'll want to test it for edge cases like /data/dir_upgrade/gpseg0 -> /data/dir_upgrade_upgrade/gpseg0, of course, but those should do it.
//	sql := "UPDATE gp_segment_configuration SET datadir = datadir || '_old';"
//	if cluster.Version.SemVer.Major < 6 {
//		sql = `SELECT dbid, content, role, preferred_role, mode, status,
//                       hostname, address, port, replication_port, fs.oid,
//                       fselocation
//                FROM pg_catalog.gp_segment_configuration
//                JOIN pg_catalog.pg_filespace_entry on (dbid = fsedbid)
//                JOIN pg_catalog.pg_filespace fs on (fsefsoid = fs.oid)
//                ORDER BY content, preferred_role DESC, fs.oid`
//	}
//
//	var mErr multierror.Error
//	errChan := make(chan error, 1)
//
//	go func() {
//		defer stdin.Close()
//		_, err = utils.System.WriteString(stdin, sql)
//		errChan <- err
//	}()
//
//	err = cmd.Run()
//	if err != nil {
//		mErr = *multierror.Append(&mErr, err)
//	}
//	close(errChan)
//
//	for e := range errChan {
//		mErr = *multierror.Append(&mErr, e)
//	}
//
//	return mErr.ErrorOrNil()
//}

func renameSegmentDataDirsOnDisk(agentConns []*Connection,
	cluster *utils.Cluster,
	srcFunc func(path string) string,
	dstFunc func(path string) string) error {

	wg := sync.WaitGroup{}
	errs := make(chan error, len(agentConns))

	for _, conn := range agentConns {
		wg.Add(1)

		go func(c *Connection) {
			defer wg.Done()

			segments, err := cluster.SegmentsOn(c.Hostname)
			if err != nil {
				errs <- err
				return
			}

			req := new(idl.ReconfigureDataDirRequest)
			for _, seg := range segments {
				pair := idl.RenamePair{
					Src: filepath.Dir(srcFunc(seg.DataDir)),
					Dst: filepath.Dir(dstFunc(seg.DataDir)),
				}

				req.Pair = append(req.Pair, &pair)
			}

			_, err = c.AgentClient.ReconfigureDataDirectories(context.Background(), req)
			if err != nil {
				gplog.Error("creating segment data directories on host %s: %s", c.Hostname, err.Error())
				errs <- err
			}
		}(conn)
	}

	wg.Wait()
	close(errs)

	var mErr *multierror.Error
	for err := range errs {
		mErr = multierror.Append(mErr, err)
	}

	return mErr.ErrorOrNil()
}

func updateGpperfmonConf(src, dst string) error {
	script := fmt.Sprintf(
		"sed 's@log_location = .*$@log_location = %[2]s/gpperfmon/logs@' %[1]s/conf/gpperfmon.conf > %[1]s/conf/gpperfmon.conf.updated && "+
			"mv %[1]s/conf/gpperfmon.conf %[1]s/conf/gpperfmon.conf.bak && "+
			"mv %[1]s/conf/gpperfmon.conf.updated %[1]s/conf/gpperfmon.conf",
		src, dst,
	)
	gplog.Debug("executing command: %+v", script) // TODO: Move this debug log into ExecuteLocalCommand()
	cmd := execCommand("bash", "-c", script)
	_, err := cmd.Output()
	if err != nil {
		return xerrors.Errorf("%s failed to execute sed command: %w",
			idl.Substep_FINALIZE_UPDATE_GPPERFMON_CONF, err)
	}
	return nil
}
