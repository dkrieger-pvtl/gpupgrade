package hub

import (
	"database/sql"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/utils"
)

const connectionString = "postgresql://localhost:%d/template1?gp_session_role=utility&allow_system_table_mods=true&search_path="

func WithinDbConnection(cluster *utils.Cluster, operation func(connection *sql.DB) error) (err error) {
	connURI := fmt.Sprintf(connectionString, cluster.MasterPort())
	connection, err := sql.Open("pgx", connURI)

	fmt.Printf("%v", connectionString)
	fmt.Printf("%v", connURI)

	if err != nil {
		return xerrors.Errorf("Failed to open connection to utility master: %w", err)
	}

	defer func() {
		closeErr := connection.Close()
		if closeErr != nil {
			closeErr = xerrors.Errorf("closing connection to new master db: %w", closeErr)
			err = multierror.Append(err, closeErr)
		}
	}()

	return operation(connection)
}

func WithinDbTransaction(cluster *utils.Cluster, operation func(transaction *sql.Tx) error) (err error) {
	return WithinDbConnection(cluster, func(connection *sql.DB) error {
		transaction, err := connection.Begin()
		if err != nil {
			return err
		}

		err = operation(transaction)
		if err != nil {
			return err
		}

		err = transaction.Commit()
		if err != nil {
			return nil
		}

		err = connection.Close()
		if err != nil {
			return err
		}

		return nil
	})
}
