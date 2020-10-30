//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"testing"
)

const ignoreLine = "do not care"

func TestCLIErrors(t *testing.T) {

	index := uint32(0)
	err := errors.New("low-level error")

	t.Run("prints out index and wrapped error", func(t *testing.T) {
		index++
		wrappedErr := NewCLIError(err)

		actual := wrappedErr.Error()
		expected := fmt.Sprintf("[CLIError %d] %s", index, err.Error())
		if actual != expected {
			t.Errorf("got %q, expected %q", actual, expected)
		}
	})

	// [CLIError 2] goroutine 21 [running]:
	// runtime/debug.Stack(0x10dc097, 0xc00011f6c0, 0xc0001be000)
	// 		/usr/local/Cellar/go/1.14.1/libexec/src/runtime/debug/stack.go:24 +0x9d
	// github.com/greenplum-db/gpupgrade/cli.NewCLIError(0x16f8860, 0xc0001196f0, 0xc000044620)
	// 		/Users/dkrieger/go/src/github.com/greenplum-db/gpupgrade/cli/errors.go:29 +0x3a
	// github.com/greenplum-db/gpupgrade/cli.TestCLIErrors.func2(0xc000157d40)
	// 		/Users/dkrieger/go/src/github.com/greenplum-db/gpupgrade/cli/errors_test.go:33 +0xde
	// testing.tRunner(0xc000157d40, 0xc000135280)
	// 		/usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:992 +0xdc
	// created by testing.(*T).Run
	//		/usr/local/Cellar/go/1.14.1/libexec/src/testing/testing.go:1043 +0x357
	t.Run("FullInfo prints out a full stack trace", func(t *testing.T) {
		index++
		wrappedErr := NewCLIError(err)

		actual := wrappedErr.FullInfo()
		expectedTemplate :=
`[CLIError %d]
runtime/debug.Stack
%s
github.com/greenplum-db/gpupgrade/cli.NewCLIError
%s
github.com/greenplum-db/gpupgrade/cli.TestCLIErrors.func
%s
testing.tRunner
%s
created by testing.(*T).Run
%s`
		expected := fmt.Sprintf(expectedTemplate, index,
			ignoreLine, ignoreLine, ignoreLine, ignoreLine, ignoreLine)

		actualScanner := bufio.NewScanner(strings.NewReader(actual))
		expectedScanner := bufio.NewScanner(strings.NewReader(expected))
		actualScanner.Split(bufio.ScanLines)
		expectedScanner.Split(bufio.ScanLines)

		for expectedScanner.Scan() {
			expectedSubstring := expectedScanner.Text()
			if ok := actualScanner.Scan(); !ok {
				t.Errorf("unexpected end of actualScanner on %s", expectedSubstring)
				break
			}
			actualLine := actualScanner.Text()

			if expectedSubstring == ignoreLine {
				continue
			}

			if !strings.Contains(actualLine, expectedSubstring) {
				t.Errorf("expected %q to contain %q", actualLine, expectedSubstring)
			}
		}
	})

}
