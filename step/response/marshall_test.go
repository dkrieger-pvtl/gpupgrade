//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package response_test

import (
	"strconv"
	"testing"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step/response"
)

func TestMessage(t *testing.T) {
	const (
		version    = "5.8.1"
		port       = 12345
		masterDir  = "/master/data/dir"
		archiveDir = "/archive/data/dir"
	)

	fields := []response.Field{
		response.UsingVersion(version),
		response.UsingMasterPort(port),
		response.UsingMasterDataDir(masterDir),
		response.UsingArchiveDir(archiveDir),
	}
	msg := response.Message(fields...)

	msgResponse := msg.GetResponse()
	if msgResponse == nil {
		t.Fatalf("expected response to not be nil")
	}
	m := msgResponse.GetData()
	if m == nil {
		t.Fatalf("expected response to not be nil")
	}

	val, ok := m[idl.ResponseKey_source_version.String()]
	if !ok {
		t.Errorf("expected field %s", idl.ResponseKey_source_version)
	}
	if val != version {
		t.Errorf("got %q want %q", val, version)
	}

	val, ok = m[idl.ResponseKey_target_port.String()]
	if !ok {
		t.Errorf("expected field %s", idl.ResponseKey_target_port)
	}
	if val != strconv.Itoa(port) {
		t.Errorf("got %q want %d", val, port)
	}

	dir, ok := m[idl.ResponseKey_target_master_data_directory.String()]
	if !ok {
		t.Errorf("expected field %s", idl.ResponseKey_target_master_data_directory)
	}
	if dir != masterDir {
		t.Errorf("got %q want %q", dir, masterDir)
	}

	dir, ok = m[idl.ResponseKey_revert_log_archive_directory.String()]
	if !ok {
		t.Errorf("expected field %s", idl.ResponseKey_revert_log_archive_directory)
	}
	if dir != archiveDir {
		t.Errorf("got %q want %q", dir, archiveDir)
	}

}
