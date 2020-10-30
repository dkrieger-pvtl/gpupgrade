//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
)

// CLIError is meant to be the lowest level error thrown by a component.
// The intent is to log the stack trace with the associated index to a log
// file, and to display to the user only the unwrapped error along with
// a reference to the stack trace.

type CLIError struct {
	internalError error
	Index         uint32
	StackTrace    string
}

var count uint32

func NewCLIError(err error) *CLIError {
	return &CLIError{
		internalError: err,
		Index:         atomic.AddUint32(&count, 1),
		StackTrace:    fmt.Sprintf("%s",debug.Stack()),
	}
}

func (e *CLIError) FullInfo() string {
	return fmt.Sprintf("[CLIError %d] %s", e.Index, e.StackTrace)
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("[CLIError %d] %s", e.Index, e.internalError.Error())
}

func (e *CLIError) Unwrap() error {
	return e.internalError
}
