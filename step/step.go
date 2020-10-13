// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/operating"
	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/utils"
	"github.com/greenplum-db/gpupgrade/utils/errorlist"
	"github.com/greenplum-db/gpupgrade/utils/stopwatch"
)

const SubstepFileName = "substep.json"

type Step struct {
	name    idl.Step
	sender  idl.MessageSender // sends substep status messages
	store   Store             // persistent substep status storage
	streams OutStreamsCloser  // writes substep stdout/err
	err     error
}

func New(name idl.Step, sender idl.MessageSender, store Store, streams OutStreamsCloser) *Step {
	return &Step{
		name:    name,
		sender:  sender,
		store:   store,
		streams: streams,
	}
}

func Begin(stateDir string, step idl.Step, sender idl.MessageSender) (*Step, error) {
	logdir, err := utils.GetLogDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(logdir, fmt.Sprintf("%s_%s.log", strings.ToLower(step.String()), operating.System.Now().Format("20060102")))
	log, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, xerrors.Errorf(`step "%s": %w`, step, err)
	}

	_, err = fmt.Fprintf(log, "\n%s in progress.\n", strings.Title(step.String()))
	if err != nil {
		log.Close()
		return nil, xerrors.Errorf(`logging step "%s": %w`, step, err)
	}

	statusPath, err := GetStatusFile(stateDir)
	if err != nil {
		return nil, xerrors.Errorf("step %q: %w", step, err)
	}

	streams := newMultiplexedStream(sender, log)

	return New(step, sender, NewFileStore(statusPath), streams), nil
}

// Returns path to status file, and if one does not exist it creates an empty
// JSON file.
func GetStatusFile(stateDir string) (path string, err error) {
	path = filepath.Join(stateDir, SubstepFileName)

	f, err := os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0600)
	if os.IsExist(err) {
		return path, nil
	}
	if err != nil {
		return "", err
	}

	defer func() {
		if cErr := f.Close(); cErr != nil {
			err = errorlist.Append(err, cErr)
		}
	}()

	// MarshallJSON requires a well-formed JSON file
	_, err = f.WriteString("{}")
	if err != nil {
		return "", err
	}

	return path, nil
}

func HasRun(step idl.Step, substep idl.Substep) (bool, error) {
	path, err := GetStatusFile(utils.GetStateDir())
	if err != nil {
		return false, xerrors.Errorf("status file: %w", err)
	}

	store := NewFileStore(path)
	status, err := store.Read(step, substep)
	if err != nil {
		return false, err
	}

	if status != idl.Status_UNKNOWN_STATUS {
		return true, nil
	}

	return false, nil
}

func (s *Step) Streams() OutStreams {
	return s.streams
}

func (s *Step) Finish() error {
	if err := s.streams.Close(); err != nil {
		return xerrors.Errorf(`step "%s": %w`, s.name, err)
	}

	return nil
}

func (s *Step) Err() error {
	return s.err
}

func (s *Step) AlwaysRun(substep idl.Substep, f func(OutStreams) error) {
	s.run(substep, f, true)
}

func (s *Step) Run(substep idl.Substep, f func(OutStreams) error) {
	s.run(substep, f, false)
}

func (s *Step) run(substep idl.Substep, f func(OutStreams) error, alwaysRun bool) {
	var err error
	defer func() {
		if err != nil {
			s.err = xerrors.Errorf(`substep "%s": %w`, substep, err)
		}
	}()

	if s.err != nil {
		return
	}

	status, err := s.store.Read(s.name, substep)
	if err != nil {
		return
	}

	if status == idl.Status_RUNNING {
		// TODO: Finalize error wording and recommended action
		err = fmt.Errorf("Found previous substep %s was running. Manual intervention needed to cleanup. Please contact support.", substep)
		s.sendStatus(substep, idl.Status_FAILED)
		return
	}

	// Only re-run substeps that are failed or pending. Do not skip substeps that must always be run.
	if status == idl.Status_COMPLETE && !alwaysRun {
		// Only send the status back to the UI; don't re-persist to the store
		s.sendStatus(substep, idl.Status_SKIPPED)
		return
	}

	timer := stopwatch.Start()
	defer func() {
		if pErr := s.printDuration(substep, timer.Stop()); pErr != nil {
			err = errorlist.Append(err, pErr)
		}
	}()

	_, err = fmt.Fprintf(s.streams.Stdout(), "\nStarting %s...\n\n", substep)
	if err != nil {
		return
	}

	err = s.write(substep, idl.Status_RUNNING)
	if err != nil {
		return
	}

	err = f(s.streams)

	switch {
	case errors.Is(err, Skip):
		// The substep has requested a manual skip; this isn't really an error.
		err = s.write(substep, idl.Status_SKIPPED)
		return

	case err != nil:
		if werr := s.write(substep, idl.Status_FAILED); werr != nil {
			err = errorlist.Append(err, werr)
		}
		return
	}

	err = s.write(substep, idl.Status_COMPLETE)
}

func (s *Step) write(substep idl.Substep, status idl.Status) error {
	storeStatus := status
	if status == idl.Status_SKIPPED {
		// Special case: we want to mark an explicitly-skipped substep COMPLETE
		// on disk.
		storeStatus = idl.Status_COMPLETE
	}

	err := s.store.Write(s.name, substep, storeStatus)
	if err != nil {
		return err
	}

	s.sendStatus(substep, status)
	return nil
}

func (s *Step) sendStatus(substep idl.Substep, status idl.Status) {
	// A stream is not guaranteed to remain connected during execution, so
	// errors are explicitly ignored.
	_ = s.sender.Send(&idl.Message{
		Contents: &idl.Message_Status{Status: &idl.SubstepStatus{
			Step:   substep,
			Status: status,
		}},
	})
}

func (s *Step) printDuration(substep idl.Substep, timer *stopwatch.Stopwatch) error {
	_, err := fmt.Fprintf(s.streams.Stdout(), "\n%s took %s\n\n", substep, timer.String())
	return err
}

// Skip can be returned from a Run or AlwaysRun callback to immediately mark the
// substep complete on disk and report "skipped" to the UI.
var Skip = skipErr{}

type skipErr struct{}

func (s skipErr) Error() string { return "skipped" }
