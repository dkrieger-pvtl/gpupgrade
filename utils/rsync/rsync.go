// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package rsync

import (
	"os/exec"

	"github.com/greenplum-db/gp-common-go-libs/gplog"

	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/testutils/exectest"
)

var rsyncCommand = exec.Command

type optionList struct {
	srcs          []string
	useDstHost    bool
	dstHost       string
	dst           string
	options       []string
	excludedFiles []string
	useStream     bool
	stream        step.OutStreams
}

func newOptionList(opts ...Option) *optionList {
	o := new(optionList)
	for _, option := range opts {
		option(o)
	}
	return o
}

type Option func(*optionList)

func WithSources(srcs ...string) Option {
	return func(options *optionList) {
		options.srcs = append(options.srcs, srcs...)
	}
}

func WithDstHost(dstHost string) Option {
	return func(options *optionList) {
		options.useDstHost = true
		options.dstHost = dstHost
	}
}

func WithDst(dst string) Option {
	return func(options *optionList) {
		options.dst = dst
	}
}

func WithOptions(opts ...string) Option {
	return func(options *optionList) {
		options.options = append(options.options, opts...)
	}
}

func WithExcludedFiles(files ...string) Option {
	return func(options *optionList) {
		for _, excludedFile := range files {
			options.excludedFiles = append(options.excludedFiles, "--exclude", excludedFile)
		}
	}
}

func WithStream(stream step.OutStreams) Option {
	return func(options *optionList) {
		options.stream = stream
		options.useStream = true
	}
}

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

//func RsyncWithoutStream(srcDir, dstHost, dst string, options, excludedFiles []string) error {
//	return RsyncWithStream(srcDir, dstHost, dst, options, excludedFiles, step.DevNullStream)
//}
//func RsyncWithStream(srcDir, dstHost, dst string, options, excludedFiles []string, stream step.OutStreams) error {
//	return Rsync([]string{srcDir}, dstHost, dst, options, excludedFiles, stream, false)
//}

//func Rsync(srcs []string, dstHost, dst string, options, excludedFiles []string, stream step.OutStreams, createSubdir bool) error {
func Rsync(options ...Option) error {
	opts := newOptionList(options...)

	//if !createSubdir {
	//	srcPaths = append(srcPaths, src+string(os.PathSeparator))
	//} else {
	//	srcPaths = append(srcPaths, src)
	//}
	var args []string

	dstPath := opts.dst
	if opts.useDstHost {
		dstPath = opts.dstHost + ":" + opts.dst
	}

	// TODO: upgrade_primaries_test.go relies on this order of arguments(!)
	args = append(args, opts.options...)
	args = append(args, opts.srcs...)
	args = append(args, dstPath)
	args = append(args, opts.excludedFiles...)

	cmd := rsyncCommand("rsync", args...)

	stream := step.BufferedStreams{}
	if opts.useStream {
		cmd.Stdout = opts.stream.Stdout()
		cmd.Stderr = opts.stream.Stderr()
	} else {
		// capture stderr if the caller does not want it, for the error message
		cmd.Stderr = stream.Stderr()
	}

	gplog.Info("running Rsync as %s", cmd.String())

	err := cmd.Run()
	if err != nil {
		return RsyncError{
			errorText: errorText(err, stream.StderrBuf.String()),
		}
	}

	return err
}

// rsync streams its interesting error message to stderr; the actual
//  error message is cryptic like "error code 12".  So, if the caller
//   did not capture stderr, place it in the error message.
func errorText(err error, stderr string) string {
	errorText := err.Error()

	if stderr != "" {
		errorText = stderr
	}

	return errorText
}
