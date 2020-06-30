//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package response_test

import (
	"testing"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step/response"
)

const (
	version    = "5.8.1"
	port       = 12345
	masterDir  = "/master/data/dir"
	archiveDir = "/archive/data/dir"
)

var execute = &idl.Response{
	Data: &idl.Response_Execute{
		Execute: &idl.ExecuteResponse{
			MasterDataDir: masterDir,
			MasterPort:    port,
		},
	},
}

var emptyMsg = &idl.Response{}

func TestMasterPort(t *testing.T) {
	t.Run("extracts port", func(t *testing.T) {
		val, err := response.MasterPort(execute)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if val != port {
			t.Errorf("got %q wanted %d", val, port)
		}
	})

	t.Run("errors with no port", func(t *testing.T) {
		port, err := response.MasterPort(emptyMsg)
		if err == nil {
			t.Errorf("expected error")
		}
		if port != 0 {
			t.Errorf("got %q wanted %q", port, "")
		}
	})
}
