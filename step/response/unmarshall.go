// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"errors"
	"os"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/idl"
)

func MasterPort(data map[string]string) (string, error) {
	version, ok := data[idl.ResponseKey_target_port.String()]
	if !ok {
		return "", errors.New("target port field not in data")
	}
	return version, nil
}

func SourceVersion(data map[string]string) (string, error) {
	version, ok := data[idl.ResponseKey_source_version.String()]
	if !ok {
		return "", errors.New("source version field not in data")
	}
	return version, nil
}

func ArchiveDir(data map[string]string) (string, error) {
	return dir(data, idl.ResponseKey_revert_log_archive_directory.String())
}

func MasterDataDir(data map[string]string) (string, error) {
	return dir(data, idl.ResponseKey_target_master_data_directory.String())
}

func dir(data map[string]string, key string) (string, error) {
	dir, ok := data[key]
	if !ok {
		return "", xerrors.Errorf("directory for field %s does not exist", key)
	}

	// make sure the returned datadir is actually a directory
	f, err := os.Stat(dir)
	if err != nil {
		return "", xerrors.Errorf("bad returned master data directory: %w", err)
	}
	if !f.IsDir() {
		return "", xerrors.Errorf("returned master data directory is not a directory: %s", dir)
	}

	return dir, nil
}
