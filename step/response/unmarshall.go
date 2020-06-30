// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"os"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/idl"
)

func MasterPort(response *idl.Response) (int, error) {

	port := 0

	if (response == nil) || (response.Data == nil) {
		return 0, xerrors.Errorf("returned response does not contain port: %#v", response)
	}

	switch x := response.Data.(type) {
	case *idl.Response_Execute:
		port = int(x.Execute.MasterPort)
	case *idl.Response_Finalize:
		port = int(x.Finalize.MasterPort)
	default:
		return 0, xerrors.Errorf("returned response does not contain port: %#v", response)
	}

	return port, nil
}

func MasterDataDir(response *idl.Response) (string, error) {

	dir := ""

	if (response == nil) || (response.Data == nil) {
		return "", xerrors.Errorf("returned response does not contain master data dir: %#v", response)
	}

	switch x := response.Data.(type) {
	case *idl.Response_Execute:
		dir = x.Execute.MasterDataDir
	case *idl.Response_Finalize:
		dir = x.Finalize.MasterDataDir
	default:
		return "", xerrors.Errorf("returned response does not contain master data dir: %#v", response)
	}

	return dir, checkDir(dir)
}

func ArchiveDir(response *idl.Response) (string, error) {

	dir := ""

	if (response == nil) || (response.Data == nil) {
		return "", xerrors.Errorf("returned response does not contain archive log dir: %#v", response)
	}

	switch x := response.Data.(type) {
	case *idl.Response_Revert:
		dir = x.Revert.ArchiveLogDir
	default:
		return "", xerrors.Errorf("returned response does not contain archive log dir: %#v", response)
	}

	return dir, checkDir(dir)
}

func SourceVersion(response *idl.Response) (string, error) {

	version := ""

	if (response == nil) || (response.Data == nil) {
		return "", xerrors.Errorf("returned response does not contain source version: %#v", response)
	}

	switch x := response.Data.(type) {
	case *idl.Response_Revert:
		version = x.Revert.SourceVersion
	default:
		return "", xerrors.Errorf("returned response does not contain source version: %#v", response)
	}

	return version, nil
}

func checkDir(dir string) error {

	// make sure the returned datadir is actually a directory
	f, err := os.Stat(dir)
	if err != nil {
		return xerrors.Errorf("bad returned master data directory: %w", err)
	}
	if !f.IsDir() {
		return xerrors.Errorf("returned master data directory is not a directory: %s", dir)
	}

	return nil
}
