//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package commanders

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/cli"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/greenplum-db/gpupgrade/utils/errorlist"
	"github.com/greenplum-db/gpupgrade/utils/stopwatch"
)

const StepFileName = "steps.json"

type CLIStep struct {
	stepName      string
	step          idl.Step
	streams       *step.BufferedStreams
	verbose       bool
	timer         *stopwatch.Stopwatch
	lastSubstep   idl.Substep
	suggestRevert bool
	err           error
}

func NewStep(step idl.Step, streams *step.BufferedStreams, verbose bool) (*CLIStep, error) {
	var err error
	stepName := strings.Title(strings.ToLower(step.String()))

	err = ValidateStep(step)
	if err != nil {
		return nil, err
	}

	fmt.Println()
	fmt.Println(stepName + " in progress.")
	fmt.Println()

	st := &CLIStep{
		stepName:      stepName,
		step:          step,
		streams:       streams,
		verbose:       verbose,
		timer:         stopwatch.Start(),
		suggestRevert: true,
	}

	// For Write to succeed the state directory needs to have been created
	// which has not yet been done when NewStep is called for Initialize.
	if step != idl.Step_INITIALIZE {
		err = Write(step, idl.Status_RUNNING)
	}

	return st, err
}

func (s *CLIStep) Err() error {
	return s.err
}

func (s *CLIStep) RunHubSubstep(f func(streams step.OutStreams) error) {
	if s.err != nil {
		return
	}

	err := f(s.streams)
	if err != nil {
		if errors.Is(err, step.Skip) {
			return
		}

		if wErr := Write(s.step, idl.Status_FAILED); wErr != nil {
			s.err = errorlist.Append(s.err, wErr)
		}

		s.err = err
	}
}

func (s *CLIStep) RunInternalSubstep(f func() error) {
	if s.err != nil {
		return
	}

	err := f()
	if err != nil {
		if errors.Is(err, step.Skip) {
			return
		}

		if wErr := Write(s.step, idl.Status_FAILED); wErr != nil {
			s.err = errorlist.Append(s.err, wErr)
		}

		s.err = err
	}
}

func (s *CLIStep) RunCLISubstep(substep idl.Substep, f func(streams step.OutStreams) error) {
	var err error
	defer func() {
		if err != nil {
			s.err = xerrors.Errorf("substep %q: %w", substep, err)

			// If deleting the state directory on the master host failed we
			// cannot write status failed since the state directory may not exist.
			if substep == idl.Substep_DELETE_MASTER_STATEDIR {
				return
			}

			if wErr := Write(s.step, idl.Status_FAILED); wErr != nil {
				s.err = errorlist.Append(s.err, wErr)
			}
		}
	}()

	if s.err != nil {
		return
	}

	substepTimer := stopwatch.Start()
	defer func() {
		logDuration(substep.String(), s.verbose, substepTimer.Stop())
	}()

	s.printStatus(substep, idl.Status_RUNNING)

	err = f(s.streams)
	if s.verbose {
		fmt.Println() // Reset the cursor so verbose output does not run into the status.

		_, wErr := s.streams.StdoutBuf.WriteTo(os.Stdout)
		if wErr != nil {
			err = errorlist.Append(err, xerrors.Errorf("writing stdout: %w", wErr))
		}

		_, wErr = s.streams.StderrBuf.WriteTo(os.Stderr)
		if wErr != nil {
			err = errorlist.Append(err, xerrors.Errorf("writing stderr: %w", wErr))
		}
	}

	if err != nil {
		status := idl.Status_FAILED

		if errors.Is(err, step.Skip) {
			status = idl.Status_SKIPPED
			err = nil
		}

		s.printStatus(substep, status)
		return
	}

	s.printStatus(substep, idl.Status_COMPLETE)
}

func (s *CLIStep) SetNextActions(suggestRevert bool) {
	s.suggestRevert = suggestRevert
}

func (s *CLIStep) Complete(completedText string) error {
	logDuration(s.stepName, s.verbose, s.timer.Stop())

	// After Finalize and Revert have completed the state directory no longer
	// exists. Thus, the status cannot be updated.
	if s.step != idl.Step_FINALIZE && s.step != idl.Step_REVERT {
		if wErr := Write(s.step, idl.Status_COMPLETE); wErr != nil {
			s.err = errorlist.Append(s.err, wErr)
		}
	}

	if s.Err() != nil {
		fmt.Println()
		return cli.NewNextActions(s.Err(), strings.ToLower(s.stepName), s.suggestRevert)
	}

	fmt.Println(completedText)
	return nil
}

func (s *CLIStep) printStatus(substep idl.Substep, status idl.Status) {
	if substep == s.lastSubstep {
		// For the same substep reset the cursor to overwrite the current status.
		fmt.Print("\r")
	}

	text := SubstepDescriptions[substep]
	fmt.Print(Format(text.OutputText, status))

	// Reset the cursor if the final status has been written. This prevents the
	// status from a hub step from being on the same line as a CLI step.
	if status != idl.Status_RUNNING {
		fmt.Println()
	}

	s.lastSubstep = substep
}

func logDuration(operation string, verbose bool, timer *stopwatch.Stopwatch) {
	msg := operation + " took " + timer.String()
	if verbose {
		fmt.Println(msg)
		fmt.Println()
	}
	gplog.Debug(msg)
}

type preconditions struct {
	notStarted idl.Step
	started    idl.Step
	completed  idl.Step
}

// TODO: disallow initialize/execute/finalize once revert has started?
var validate = map[idl.Step]preconditions{
	idl.Step_INITIALIZE: {
		notStarted: idl.Step_EXECUTE,
	},
	idl.Step_EXECUTE: {
		notStarted: idl.Step_FINALIZE,
		completed:  idl.Step_INITIALIZE,
	},
	idl.Step_FINALIZE: {
		completed: idl.Step_EXECUTE,
	},
	idl.Step_REVERT: {
		notStarted: idl.Step_FINALIZE,
		started:    idl.Step_INITIALIZE,
	},
}

type ValidateStepError struct {
	step       idl.Step
	conditions preconditions
	lookupErr  error
}

func NewValidateStepError(step idl.Step, conditions preconditions, lookupErr error) ValidateStepError {
	return ValidateStepError{step: step, conditions: conditions, lookupErr: lookupErr}
}

func (v ValidateStepError) Error() string {
	msg := fmt.Sprintf("Step %s cannot be run:", v.step.String())
	if v.lookupErr != nil {
		return fmt.Sprintf("%s\nCannot determine if step is running: %s", msg, v.lookupErr.Error())
	}

	if v.conditions.notStarted != idl.Step_UNKNOWN_STEP {
		msg += fmt.Sprintf("\nstep %s must not have started", v.conditions.notStarted.String())
	}

	if v.conditions.started != idl.Step_UNKNOWN_STEP {
		msg += fmt.Sprintf("\nstep %s must have started", v.conditions.started.String())
	}

	if v.conditions.completed != idl.Step_UNKNOWN_STEP {
		msg += fmt.Sprintf("\nstep %s must have completed", v.conditions.completed.String())
	}

	return msg
}

func (v ValidateStepError) Unwrap() error {
	return v.lookupErr
}

// TODO: return an error type that provides the reasons the current step cannot run
// TODO: consider how to fix case of file not existing before INTIALIZ....
func ValidateStep(stepName idl.Step) error {
	conditions, ok := validate[stepName]
	if !ok {
		return NewValidateStepError(stepName, conditions, fmt.Errorf("internal error: cannot lookup step name"))
	}

	if conditions.notStarted != idl.Step_UNKNOWN_STEP {
		hasStarted, err := HasStepStarted(conditions.notStarted)
		if stepName == idl.Step_INITIALIZE {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return nil
			}
		}
		if hasStarted || err != nil {
			return NewValidateStepError(stepName, conditions, err)
		}
	}

	if conditions.started != idl.Step_UNKNOWN_STEP {
		hasStarted, err := HasStepStarted(conditions.started)
		if stepName == idl.Step_INITIALIZE {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return nil
			}
		}
		if !hasStarted || err != nil {
			return NewValidateStepError(stepName, conditions, err)
		}
	}

	if conditions.completed != idl.Step_UNKNOWN_STEP {
		hasCompleted, err := HasStepCompleted(conditions.completed)
		if stepName == idl.Step_INITIALIZE {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				return nil
			}
		}
		if !hasCompleted || err != nil {
			return NewValidateStepError(stepName, conditions, err)
		}
	}

	return nil
}

func Write(stepName idl.Step, status idl.Status) error {
	path, err := utils.GetJSONFile(utils.GetStateDir(), StepFileName)
	if err != nil {
		return xerrors.Errorf("getting %q file: %w", StepFileName, err)
	}

	store := step.NewFileStore(path)

	if status == idl.Status_SKIPPED {
		// Special case: we want to mark an explicitly-skipped substep COMPLETE
		// on disk.
		status = idl.Status_COMPLETE
	}

	err = store.Write(stepName, idl.Substep_INTERNAL_STEP_STATUS, status)
	if err != nil {
		return err
	}

	return nil
}

func HasStepStarted(stepName idl.Step) (bool, error) {
	return HasStatus(stepName, func(status idl.Status) bool {
		return status != idl.Status_UNKNOWN_STATUS
	})
}

func HasStepCompleted(stepName idl.Step) (bool, error) {
	return HasStatus(stepName, func(status idl.Status) bool {
		return status == idl.Status_COMPLETE
	})
}

func HasStatus(stepName idl.Step, check func(status idl.Status) bool) (bool, error) {
	path, err := utils.GetJSONFile(utils.GetStateDir(), StepFileName)
	if err != nil {
		return false, xerrors.Errorf("getting %q file: %w", StepFileName, err)
	}

	store := step.NewFileStore(path)

	status, err := store.Read(stepName, idl.Substep_INTERNAL_STEP_STATUS)
	if err != nil {
		return false, err
	}

	return check(status), nil
}
