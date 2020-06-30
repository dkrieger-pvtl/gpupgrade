//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package response_test

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step/response"
	"github.com/greenplum-db/gpupgrade/testutils"
)

const (
	version    = "5.8.1"
	port       = 12345
	masterDir  = "/master/data/dir"
	archiveDir = "/archive/data/dir"
)

var dataMap = map[string]string{
	idl.ResponseKey_source_version.String():               version,
	idl.ResponseKey_target_port.String():                  strconv.Itoa(port),
	idl.ResponseKey_target_master_data_directory.String(): masterDir,
	idl.ResponseKey_revert_log_archive_directory.String(): archiveDir,
}

var emptyMap map[string]string

func TestMasterPort(t *testing.T) {
	t.Run("extracts port", func(t *testing.T) {
		val, err := response.MasterPort(dataMap)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if val != strconv.Itoa(port) {
			t.Errorf("got %q wanted %d", val, port)
		}
	})

	t.Run("errors with no port", func(t *testing.T) {
		port, err := response.MasterPort(emptyMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if port != "" {
			t.Errorf("got %q wanted %q", port, "")
		}
	})
}

func TestSourceVersion(t *testing.T) {
	t.Run("extracts source version", func(t *testing.T) {
		val, err := response.SourceVersion(dataMap)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if val != version {
			t.Errorf("got %q wanted %q", val, version)
		}
	})

	t.Run("errors with no version", func(t *testing.T) {
		version, err := response.MasterPort(emptyMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if version != "" {
			t.Errorf("got %q wanted %q", version, "")
		}
	})
}

func TestArchiveDir(t *testing.T) {
	tmpDir := testutils.GetTempDir(t, "")
	defer testutils.MustRemoveAll(t, tmpDir)
	file := filepath.Join(tmpDir, "file.txt")
	testutils.MustWriteToFile(t, file, "")

	m := map[string]string{
		idl.ResponseKey_revert_log_archive_directory.String(): tmpDir,
	}

	t.Run("extracts archiveDir", func(t *testing.T) {
		dir, err := response.ArchiveDir(m)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if dir != tmpDir {
			t.Errorf("got %q wanted %q", dir, tmpDir)
		}
	})

	t.Run("errors with no archiveDir", func(t *testing.T) {
		dir, err := response.ArchiveDir(emptyMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})

	t.Run("errors with bad archiveDir", func(t *testing.T) {
		dir, err := response.ArchiveDir(dataMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})

	t.Run("errors with file as archiveDir", func(t *testing.T) {
		m[idl.ResponseKey_revert_log_archive_directory.String()] = file

		dir, err := response.ArchiveDir(m)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})
}

func TestMasterDataDir(t *testing.T) {
	tmpDir := testutils.GetTempDir(t, "")
	defer testutils.MustRemoveAll(t, tmpDir)
	file := filepath.Join(tmpDir, "file.txt")
	testutils.MustWriteToFile(t, file, "")

	m := map[string]string{
		idl.ResponseKey_target_master_data_directory.String(): tmpDir,
	}

	t.Run("extracts masterDataDir", func(t *testing.T) {
		dir, err := response.MasterDataDir(m)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		if dir != tmpDir {
			t.Errorf("got %q wanted %q", dir, tmpDir)
		}
	})

	t.Run("errors with no masterDataDir", func(t *testing.T) {
		dir, err := response.MasterDataDir(emptyMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})

	t.Run("errors with bad masterDataDir", func(t *testing.T) {
		dir, err := response.MasterDataDir(dataMap)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})

	t.Run("errors with file as masterDataDir", func(t *testing.T) {
		m[idl.ResponseKey_target_master_data_directory.String()] = file

		dir, err := response.MasterDataDir(m)
		if err == nil {
			t.Errorf("expected error")
		}
		if dir != "" {
			t.Errorf("got %q wanted %q", dir, "")
		}
	})
}
