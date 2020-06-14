// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package rsync

import (
	"os"
	"os/exec"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/testutils/exectest"
)

var rsyncCommand = exec.Command

// TODO: this function should be test-only but needs to be used in other
//  components that use Rsync in their implementation.
func SetRsyncCommand(command exectest.Command) {
	rsyncCommand = command
}
func ResetRsyncCommand() {
	rsyncCommand = exec.Command
}

type RsyncError struct {
	errorText string
}

func (e RsyncError) Error() string {
	return e.errorText
}

func RsyncWithoutStream(srcDir, dstHost, dst string, options, excludedFiles []string) error {
	return RsyncWithStream(srcDir, dstHost, dst, options, excludedFiles, step.DevNullStream)
}
func RsyncWithStream(srcDir, dstHost, dst string, options, excludedFiles []string, stream step.OutStreams) error {
	return Rsync([]string{srcDir}, dstHost, dst, options, excludedFiles, stream, false)
}

// rsync src1/ src2/ host:dest --option1 --option2 --exclude foo.txt
func Rsync(srcs []string, dstHost, dst string, options, excludedFiles []string, stream step.OutStreams, createSubdir bool) error {
	var srcPaths []string
	for _, src := range srcs {
		if !createSubdir {
			srcPaths = append(srcPaths, src+string(os.PathSeparator))
		} else {
			srcPaths = append(srcPaths, src)
		}
	}
	dstPath := dst
	if dstHost != "" {
		dstPath = dstHost + ":" + dst
	}

	// TODO: upgrade_primaries_test.go relies on this order of arguments(!)
	var args []string
	args = append(args, options...)
	args = append(args, srcPaths...)
	args = append(args, dstPath)
	args = append(args, exclusionOpts(excludedFiles)...)

	cmd := rsyncCommand("rsync", args...)

	//bufStdout := bytes.Buffer{}
	//bufStderr := bytes.Buffer{}
	//bufStream := step.NewBufStream(&bufStdout, &bufStderr)
	//tee := step.NewTeeStream(stream, bufStream)
	//
	//cmd.Stdout = tee.Stdout()
	//cmd.Stderr = tee.Stderr()
	cmd.Stdout = stream.Stdout()
	cmd.Stderr = stream.Stderr()

	gplog.Info("running Rsync as %s", cmd.String())

	return cmd.Run()
	//if err != nil {
	//	return RsyncError{
	//		errorText: errorText(err, bufStderr.String()),
	//	}
	//}
	//
	//return nil
}

// errorText provides text for an RsyncError that makes sense
//  to a user.  If the Rsync() fails to run at all, the relevant
//  text will be in the error returned from the cmd itself.  If
//  Rsync() fails during execution, the cmd will return a type
//  exec.ExitError but the relevant text is in the stderr from the
//  command.
//func errorText(err error, stderr string) string {
//	errorText := err.Error()
//
//	var exitError *exec.ExitError
//	if xerrors.As(err, &exitError) {
//		errorText = stderr
//		if stderr == "" {
//			errorText = string(exitError.Stderr)
//		}
//	}
//
//	return errorText
//}

func exclusionOpts(excludedFiles []string) []string {
	var exclusions []string
	for _, excludedFile := range excludedFiles {
		exclusions = append(exclusions, "--exclude", excludedFile)
	}
	return exclusions
}
