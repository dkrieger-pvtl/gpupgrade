//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package commanders_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/greenplum-db/gpupgrade/cli"
	"github.com/greenplum-db/gpupgrade/cli/commanders"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/testutils"
	"github.com/greenplum-db/gpupgrade/utils"
)

func TestStepStore(t *testing.T) {
	stateDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(stateDir); err != nil {
			t.Errorf("removing temp directory: %v", err)
		}
	}()

	resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", stateDir)
	defer resetEnv()

	store, err := commanders.NewStepStore()
	if err != nil {
		t.Fatalf("NewStepStore failed: %v", err)
	}

	t.Run("write persists the step status", func(t *testing.T) {
		expected := idl.Status_RUNNING
		err := store.Write(idl.Step_INITIALIZE, expected)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		status, err := store.Read(idl.Step_INITIALIZE)
		if err != nil {
			t.Errorf("Read failed %#v", err)
		}

		if status != expected {
			t.Errorf("got stauts %q want %q", status, expected)
		}
	})

	t.Run("cannot create a new step store if state directory does not exist", func(t *testing.T) {
		resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", "/does/not/exist")
		defer resetEnv()

		store, err := commanders.NewStepStore()
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			t.Errorf("got %T, want %T", err, pathErr)
		}

		expected := &commanders.StepStore{}
		if !reflect.DeepEqual(store, expected) {
			t.Errorf("got %v want %v", store, expected)
		}
	})

	t.Run("write errors and read errors with unknown status when failing to get the steps status file", func(t *testing.T) {
		stateDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(stateDir); err != nil {
				t.Errorf("removing temp directory: %v", err)
			}
		}()

		resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", stateDir)
		defer resetEnv()

		store, err := commanders.NewStepStore()
		if err != nil {
			t.Fatalf("NewStepStore failed: %v", err)
		}

		// remove state directory
		if err := os.RemoveAll(stateDir); err != nil {
			t.Fatalf("removing temp state directory: %v", err)
		}

		err = store.Write(idl.Step_INITIALIZE, idl.Status_RUNNING)
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			t.Errorf("returned error type %T want %T", err, pathErr)
		}

		status, err := store.Read(idl.Step_INITIALIZE)
		if !errors.As(err, &pathErr) {
			t.Errorf("returned error type %T want %T", err, pathErr)
		}

		expected := idl.Status_UNKNOWN_STATUS
		if status != expected {
			t.Errorf("got stauts %q want %q", status, expected)
		}
	})

	t.Run("HasStatus errors with false when failing to get the steps status file", func(t *testing.T) {
		stateDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(stateDir); err != nil {
				t.Errorf("removing temp directory: %v", err)
			}
		}()

		resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", stateDir)
		defer resetEnv()

		store, err := commanders.NewStepStore()
		if err != nil {
			t.Fatalf("NewStepStore failed: %v", err)
		}

		// remove state directory
		if err := os.RemoveAll(stateDir); err != nil {
			t.Fatalf("removing temp state directory: %v", err)
		}

		called := false
		check := func(status idl.Status) bool {
			called = true
			return true
		}

		hasStatus, err := store.HasStatus(idl.Step_INITIALIZE, check)
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			t.Errorf("returned error type %T want %T", err, pathErr)
		}

		if hasStatus {
			t.Error("expected hasStatus to be false")
		}

		if called {
			t.Error("expected check function to not be called")
		}
	})

	t.Run("HasStepStarted returns true if a step's status is running, complete, or failed", func(t *testing.T) {
		statuses := []idl.Status{idl.Status_RUNNING, idl.Status_COMPLETE, idl.Status_FAILED}
		for _, status := range statuses {
			err := store.Write(idl.Step_INITIALIZE, status)
			if err != nil {
				t.Errorf("Write failed %#v", err)
			}

			started, err := store.HasStepStarted(idl.Step_INITIALIZE)
			if err != nil {
				t.Errorf("HasStepStarted failed %#v", err)
			}

			if !started {
				t.Error("expected step to have been started")
			}
		}
	})

	t.Run("HasStepStarted returns false if a step has not started", func(t *testing.T) {
		started, err := store.HasStepStarted(idl.Step_UNKNOWN_STEP)
		if err != nil {
			t.Errorf("HasStepStarted failed %#v", err)
		}

		if started {
			t.Error("expected step to have not been started")
		}
	})

	t.Run("HasStepCompleted returns true if a step's status is complete", func(t *testing.T) {
		err := store.Write(idl.Step_INITIALIZE, idl.Status_COMPLETE)
		if err != nil {
			t.Errorf("Write failed %#v", err)
		}

		completed, err := store.HasStepCompleted(idl.Step_INITIALIZE)
		if err != nil {
			t.Errorf("HasStepCompleted failed %#v", err)
		}

		if !completed {
			t.Error("expected step to have been completed")
		}
	})

	t.Run("HasStepCompleted returns false if a step's status is not complete", func(t *testing.T) {
		statuses := []idl.Status{idl.Status_RUNNING, idl.Status_FAILED, idl.Status_SKIPPED, idl.Status_UNKNOWN_STATUS}
		for _, status := range statuses {
			err := store.Write(idl.Step_INITIALIZE, status)
			if err != nil {
				t.Errorf("Write failed %#v", err)
			}

			completed, err := store.HasStepCompleted(idl.Step_INITIALIZE)
			if err != nil {
				t.Errorf("HasStepCompleted failed %#v", err)
			}

			if completed {
				t.Error("expected step to have not been completed")
			}
		}
	})
}

func TestValidateStep(t *testing.T) {
	stateDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(stateDir); err != nil {
			t.Errorf("removing temp directory: %v", err)
		}
	}()

	resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", stateDir)
	defer resetEnv()

	store, err := commanders.NewStepStore()
	if err != nil {
		t.Fatalf("NewStepStore failed: %v", err)
	}

	t.Run("fails when execute, finalize, and revert are run and initialize has not completed", func(t *testing.T) {
		clearStore(t)

		steps := []idl.Step{
			idl.Step_EXECUTE,
			idl.Step_FINALIZE,
			idl.Step_REVERT,
		}

		for _, s := range steps {
			err := store.ValidateStep(s)
			var nextActionsErr cli.NextActions
			if !errors.As(err, &nextActionsErr) {
				t.Errorf("got %T, want %T", err, nextActionsErr)
			}

			if nextActionsErr.NextAction != commanders.NextActionRunInitialize {
				t.Errorf("got %q want %q", nextActionsErr.NextAction, commanders.NextActionRunInitialize)
			}
		}
	})

	t.Run("fails when initialize is run and execute has already started", func(t *testing.T) {
		clearStore(t)

		mustWriteStatus(t, store, idl.Step_EXECUTE, idl.Status_RUNNING)

		err = store.ValidateStep(idl.Step_INITIALIZE)
		var nextActionsErr cli.NextActions
		if !errors.As(err, &nextActionsErr) {
			t.Errorf("got %T, want %T", err, nextActionsErr)
		}

		if nextActionsErr.NextAction != commanders.NextActionRunExecute {
			t.Errorf("got %q want %q", nextActionsErr.NextAction, commanders.NextActionRunFinalize)
		}
	})

	t.Run("fails when finalize is run and execute has not completed", func(t *testing.T) {
		clearStore(t)

		mustWriteStatus(t, store, idl.Step_INITIALIZE, idl.Status_COMPLETE)

		err = store.ValidateStep(idl.Step_FINALIZE)
		var nextActionsErr cli.NextActions
		if !errors.As(err, &nextActionsErr) {
			t.Errorf("got %T, want %T", err, nextActionsErr)
		}

		if nextActionsErr.NextAction != commanders.NextActionRunExecute {
			t.Errorf("got %q want %q", nextActionsErr.NextAction, commanders.NextActionRunExecute)
		}
	})

	t.Run("fails when revert is run and finalize has already been started", func(t *testing.T) {
		clearStore(t)

		mustWriteStatus(t, store, idl.Step_INITIALIZE, idl.Status_COMPLETE)
		mustWriteStatus(t, store, idl.Step_EXECUTE, idl.Status_COMPLETE)
		mustWriteStatus(t, store, idl.Step_FINALIZE, idl.Status_RUNNING)

		err = store.ValidateStep(idl.Step_REVERT)
		var nextActionsErr cli.NextActions
		if !errors.As(err, &nextActionsErr) {
			t.Errorf("got %T, want %T", err, nextActionsErr)
		}

		if nextActionsErr.NextAction != commanders.NextActionCompleteFinalize {
			t.Errorf("got %q want %q", nextActionsErr.NextAction, commanders.NextActionCompleteFinalize)
		}
	})

	t.Run("fails when initialize, execute are run after finalize has started", func(t *testing.T) {
		clearStore(t)

		mustWriteStatus(t, store, idl.Step_FINALIZE, idl.Status_RUNNING)

		steps := []idl.Step{
			idl.Step_INITIALIZE,
			idl.Step_EXECUTE,
		}

		for _, s := range steps {
			err := store.ValidateStep(s)
			var nextActionsErr cli.NextActions
			if !errors.As(err, &nextActionsErr) {
				t.Errorf("got %T, want %T", err, nextActionsErr)
			}

			if nextActionsErr.NextAction != commanders.NextActionRunFinalize {
				t.Errorf("got %q want %q", nextActionsErr.NextAction, commanders.NextActionRunFinalize)
			}
		}
	})
}

func clearStore(t *testing.T) {
	t.Helper()

	path := filepath.Join(utils.GetStateDir(), commanders.StepsFileName)
	testutils.MustWriteToFile(t, path, "{}")
}

func mustWriteStatus(t *testing.T, store *commanders.StepStore, step idl.Step, status idl.Status) {
	t.Helper()

	err := store.Write(step, status)
	if err != nil {
		t.Errorf("store.Write returned error %+v", err)
	}
}
