//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package response_test

import (
	"reflect"
	"testing"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step/response"
)

func TestExecuteMessage(t *testing.T) {
	expected := &idl.Message{
		Contents: &idl.Message_Response{
			Response: &idl.Response{
				Data: &idl.Response_Execute{
					Execute: &idl.ExecuteResponse{
						MasterDataDir: masterDir,
						MasterPort:    int32(port),
					},
				},
			},
		},
	}

	actual := response.ExecuteMesssage(masterDir, port)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v want %v", actual, expected)
	}
}
