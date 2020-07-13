// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/upgrade"
	"github.com/greenplum-db/gpupgrade/utils"
)

var deleteDirectories = upgrade.DeleteDirectories

func (s *Server) DeleteStateDirectory(ctx context.Context, in *idl.DeleteStateDirectoryRequest) (*idl.DeleteStateDirectoryReply, error) {
	gplog.Info("got a request to delete the state directory from the hub")

	hostname, err := utils.System.Hostname()
	if err != nil {
		return &idl.DeleteStateDirectoryReply{}, err
	}

	err = deleteDirectories([]string{s.conf.StateDir}, upgrade.StateDirectoryFiles, hostname, step.DevNullStream)
	return &idl.DeleteStateDirectoryReply{}, err
}

func (s *Server) DeleteDataDirectories(ctx context.Context, in *idl.DeleteDataDirectoriesRequest) (*idl.DeleteDataDirectoriesReply, error) {
	gplog.Info("got a request to delete data directories from the hub")

	hostname, err := utils.System.Hostname()
	if err != nil {
		return &idl.DeleteDataDirectoriesReply{}, err
	}

	err = deleteDirectories(in.Datadirs, upgrade.PostgresFiles, hostname, step.DevNullStream)
	return &idl.DeleteDataDirectoriesReply{}, err
}

func (s *Server) DeleteTablespaceDirectories(ctx context.Context, in *idl.DeleteTablespaceRequest) (*idl.DeleteTablespaceReply, error) {
	gplog.Info("got a request to delete tablespace directories from the hub")

	hostname, err := utils.System.Hostname()
	if err != nil {
		return &idl.DeleteTablespaceReply{}, err
	}

	err = deleteDirectories(in.GetDirs(), []string{}, hostname, step.DevNullStream)
	if err != nil {
		return &idl.DeleteTablespaceReply{}, err
	}

	// For Tablespaces we have to delete the top level diretory IFF its empty.
	// there is a race between checking if the directory is empty and removing it...
	for _, dir := range in.GetDirs() {
		parent := filepath.Dir(dir)

		entries, err := ioutil.ReadDir(parent)
		if err != nil {
			return &idl.DeleteTablespaceReply{}, err
		}

		// directory is not empty, so it is one we did not create. We can't delete it.
		if len(entries) > 0 {
			return &idl.DeleteTablespaceReply{}, nil
		}

		// Directory is empty, so it is one we created. So deleted it.
		if err := os.Remove(parent); err != nil {
			return &idl.DeleteTablespaceReply{}, err
		}
	}

	return &idl.DeleteTablespaceReply{}, err
}
