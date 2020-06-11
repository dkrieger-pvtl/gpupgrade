// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"os"
	"os/exec"

	"golang.org/x/xerrors"
)

var rsyncCommand = exec.Command

type RsyncError struct {
	errorText string
}

func (e RsyncError) Error() string {
	return e.errorText
}

// todo: capture stdout...
func Rsync(src, dstHost, dst string, options []string, excludedFiles []string) error {
	dstFull := dst
	if dstHost != "" {
		dstFull = dstHost + ":" + dst
	}

	// TODO: upgrade_primaries_test.go relies on this order of arguments(!)
	var args []string
	args = append(args, options...)
	args = append(args, src+string(os.PathSeparator)) // the trailing path separator is critical for rsync
	args = append(args, dstFull)
	args = append(args, makeExclusionList(excludedFiles)...)

	if _, err := rsyncCommand("rsync", args...).Output(); err != nil {
		return RsyncError{
			errorText: extractTextFromError(err),
		}
	}

	return nil
}

func extractTextFromError(err error) string {
	var exitError *exec.ExitError
	errorText := err.Error()

	if xerrors.As(err, &exitError) {
		errorText = string(exitError.Stderr)
	}
	return errorText
}

func makeExclusionList(excludedFiles []string) []string {
	var exclusions []string
	for _, excludedFile := range excludedFiles {
		exclusions = append(exclusions, "--exclude", excludedFile)
	}
	return exclusions
}
