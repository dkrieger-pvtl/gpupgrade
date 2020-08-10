module github.com/greenplum-db/gpupgrade

go 1.14

require (
	github.com/DATA-DOG/go-sqlmock v1.4.0
	github.com/blang/semver/v4 v4.0.0
	github.com/cloudfoundry/gosigar v1.1.0
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.1
	github.com/google/renameio v0.1.0
	github.com/greenplum-db/gp-common-go-libs v1.0.4
	github.com/hashicorp/go-multierror v1.0.0
	github.com/jackc/pgx v3.2.0+incompatible
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/lib/pq v1.3.0
	github.com/onsi/gomega v1.7.1
	github.com/pkg/errors v0.8.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/sys v0.0.0-20200806125547-5acd03effb82
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/grpc v1.27.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v0.0.0-20200810225334-2983360ff4e7 // indirect
	google.golang.org/protobuf v1.25.0
)
