package commanders

import (
	"github.com/greenplum-db/gpupgrade/idl"
)

type VersionChecker struct {
	client idl.CliToHubClient
}

func NewVersionChecker(client idl.CliToHubClient) VersionChecker {
	return VersionChecker{
		client: client,
	}
}

func (req VersionChecker) Execute() (err error) {
	s := Substep("Checking version compatibility...")
	defer s.Finish(&err)

	// TODO: this should really just be checking the versions of greenplum as
	//  reported by "postgres --gp-version" on each host
	//resp, err := req.client.CheckVersion(context.Background(), &idl.CheckVersionRequest{})
	//if err != nil {
	//	return errors.Wrap(err, "gRPC call to hub failed")
	//}
	//if !resp.IsVersionCompatible {
	//	return errors.New("Version Compatibility Check Failed")
	//}

	return nil
}
