// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/greenplum-db/gpupgrade/idl"
)

func ExecuteMesssage(masterDataDir string, masterPort int) *idl.Message {
	return &idl.Message{
		Contents: &idl.Message_Response{
			Response: &idl.Response{
				Data: &idl.Response_Execute{
					Execute: &idl.ExecuteResponse{
						MasterDataDir: masterDataDir,
						MasterPort:    int32(masterPort),
					},
				},
			},
		},
	}
}

func FinalizeMesssage(masterDataDir string, masterPort int) *idl.Message {
	return &idl.Message{
		Contents: &idl.Message_Response{
			Response: &idl.Response{
				Data: &idl.Response_Finalize{
					Finalize: &idl.FinalizeResponse{
						MasterDataDir: masterDataDir,
						MasterPort:    int32(masterPort),
					},
				},
			},
		},
	}
}

func RevertMesssage(archiveLogDir string, sourceVersion string) *idl.Message {
	return &idl.Message{
		Contents: &idl.Message_Response{
			Response: &idl.Response{
				Data: &idl.Response_Revert{
					Revert: &idl.RevertResponse{
						ArchiveLogDir: archiveLogDir,
						SourceVersion: sourceVersion,
					},
				},
			},
		},
	}
}
