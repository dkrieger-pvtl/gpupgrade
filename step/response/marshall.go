// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"strconv"

	"github.com/greenplum-db/gpupgrade/idl"
)

func Message(opts ...Field) *idl.Message {
	fields := newField(opts...)

	data := make(map[string]string)
	data[idl.ResponseKey_source_version.String()] = fields.version
	data[idl.ResponseKey_revert_log_archive_directory.String()] = fields.archiveDir
	data[idl.ResponseKey_target_master_data_directory.String()] = fields.masterDataDir
	data[idl.ResponseKey_target_port.String()] = strconv.Itoa(fields.masterPort)

	return &idl.Message{
		Contents: &idl.Message_Response{
			Response: &idl.Response{Data: data},
		},
	}
}

type Field func(*fields)

func UsingVersion(version string) Field {
	return func(m *fields) {
		m.version = version
	}
}

func UsingMasterDataDir(path string) Field {
	return func(m *fields) {
		m.masterDataDir = path
	}
}

func UsingMasterPort(port int) Field {
	return func(m *fields) {
		m.masterPort = port
	}
}

func UsingArchiveDir(path string) Field {
	return func(m *fields) {
		m.archiveDir = path
	}
}

type fields struct {
	version       string
	masterDataDir string
	masterPort    int
	archiveDir    string
}

func newField(opts ...Field) *fields {
	m := &fields{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}
