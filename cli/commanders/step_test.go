//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package commanders_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/greenplum-db/gpupgrade/cli"
	"github.com/greenplum-db/gpupgrade/cli/commanders"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/testutils"
)

func TestSubstep(t *testing.T) {
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

	t.Run("substep status is correctly printed on success and failure", func(t *testing.T) {
		d := commanders.BufferStandardDescriptors(t)

		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		st.RunCLISubstep(idl.Substep_CHECK_DISK_SPACE, func(streams step.OutStreams) error {
			return nil
		})

		err = errors.New("error")
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return err
		})

		err = st.Complete("")
		if err == nil {
			d.Close()
			t.Errorf("want err got nil")
		}

		stdout, stderr := d.Collect()
		d.Close()
		if len(stderr) != 0 {
			t.Errorf("unexpected stderr %#v", string(stderr))
		}

		expected := "\nInitialize in progress.\n\n"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_CHECK_DISK_SPACE].OutputText, idl.Status_RUNNING) + "\r"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_CHECK_DISK_SPACE].OutputText, idl.Status_COMPLETE) + "\n"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_RUNNING) + "\r"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_FAILED) + "\n"
		expected += "\n"

		actual := string(stdout)
		if actual != expected {
			t.Errorf("output %#v want %#v", actual, expected)
			t.Logf("actual: %s", actual)
			t.Logf("expected: %s", expected)
		}
	})

	t.Run("there is no error when a hub substep is skipped", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		expected := step.Skip
		st.RunHubSubstep(func(streams step.OutStreams) error {
			return expected
		})

		err = st.Complete("")
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		if st.Err() != nil {
			t.Errorf("want err to be set to nil, got %#v", expected)
		}
	})

	t.Run("when a CLI substep is skipped its status is printed without error", func(t *testing.T) {
		d := commanders.BufferStandardDescriptors(t)

		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		skipErr := step.Skip
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return skipErr
		})

		err = st.Complete("")
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		if st.Err() != nil {
			t.Errorf("want err to be set to nil, got %#v", skipErr)
		}

		stdout, stderr := d.Collect()
		d.Close()
		if len(stderr) != 0 {
			t.Errorf("unexpected stderr %#v", string(stderr))
		}

		expected := "\nInitialize in progress.\n\n"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_RUNNING) + "\r"
		expected += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_SKIPPED) + "\n"
		expected += "\n"

		actual := string(stdout)
		if actual != expected {
			t.Errorf("output %#v want %#v", actual, expected)
			t.Logf("actual: %s", actual)
			t.Logf("expected: %s", expected)
		}
	})

	t.Run("there is no error when an internal substep is skipped", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		skipErr := step.Skip
		st.RunInternalSubstep(func() error {
			return skipErr
		})

		err = st.Complete("")
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		if st.Err() != nil {
			t.Errorf("want err to be set to nil, got %#v", skipErr)
		}
	})

	t.Run("both cli and hub substeps are not run when an internal substep errors", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		err = errors.New("error")
		st.RunInternalSubstep(func() error {
			return err
		})

		ran := false
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			ran = true
			return nil
		})

		st.RunHubSubstep(func(streams step.OutStreams) error {
			ran = true
			return nil
		})

		err = st.Complete("")
		if err == nil {
			t.Errorf("expected error")
		}

		if ran {
			t.Error("expected substep to not be run")
		}

		if st.Err() == nil {
			t.Error("expected error")
		}
	})

	t.Run("nothing is printed for internal substeps", func(t *testing.T) {
		d := commanders.BufferStandardDescriptors(t)

		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, true)
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		ran := false
		st.RunInternalSubstep(func() error {
			ran = true
			return nil
		})

		err = st.Complete("")
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		if !ran {
			d.Close()
			t.Error("expected hub substep to be run")
		}

		if st.Err() != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		expectedStdout := "\nInitialize in progress.\n\n"
		expectedStdout += "Initialize took 0s\n\n"
		expectedStdout += "\n"

		stdout, stderr := d.Collect()
		d.Close()
		actualStdout := string(stdout)
		if actualStdout != expectedStdout {
			t.Errorf("stdout %#v want %#v", actualStdout, expectedStdout)
			t.Logf("actualStdout: %s", actualStdout)
			t.Logf("expectedStdout: %s", expectedStdout)
		}

		actualStderr := string(stderr)
		expectedStderr := ""
		if actualStderr != expectedStderr {
			t.Errorf("stderr %#v want %#v", actualStdout, expectedStderr)
		}
	})

	t.Run("cli substeps are printed to stdout and stderr in verbose mode", func(t *testing.T) {
		d := commanders.BufferStandardDescriptors(t)

		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, true)
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		substepStdout := "some substep output text."
		substepStderr := "oops!"
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			os.Stdout.WriteString(substepStdout)
			os.Stderr.WriteString(substepStderr)
			return nil
		})

		err = st.Complete("")
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		expectedStdout := "\nInitialize in progress.\n\n"
		expectedStdout += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_RUNNING)
		expectedStdout += substepStdout + "\n\r"
		expectedStdout += commanders.Format(commanders.SubstepDescriptions[idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG].OutputText, idl.Status_COMPLETE) + "\n"
		expectedStdout += fmt.Sprintf("%s took", idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG)

		stdout, stderr := d.Collect()
		d.Close()
		actualStdout := string(stdout)
		// Use HasPrefix since we don't know the actualStdout step duration.
		if !strings.HasPrefix(actualStdout, expectedStdout) {
			t.Errorf("stdout %#v want %#v", actualStdout, expectedStdout)
			t.Logf("actualStdout: %s", actualStdout)
			t.Logf("expectedStdout: %s", expectedStdout)
		}

		actualStderr := string(stderr)
		if actualStderr != substepStderr {
			t.Errorf("stderr %#v want %#v", actualStdout, expectedStdout)
		}
	})

	t.Run("cli substeps are not run when there is an error", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		err = errors.New("error")
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return err
		})

		ran := false
		st.RunCLISubstep(idl.Substep_START_HUB, func(streams step.OutStreams) error {
			ran = true
			return nil
		})

		err = st.Complete("")
		if err == nil {
			t.Errorf("expected error")
		}

		if ran {
			t.Error("expected cli substep to not be run")
		}

		if st.Err() == nil {
			t.Error("expected error")
		}
	})

	t.Run("hub substeps are not run when there is an error", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		err = errors.New("error")
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return err
		})

		ran := false
		st.RunHubSubstep(func(streams step.OutStreams) error {
			ran = true
			return nil
		})

		err = st.Complete("")
		if err == nil {
			t.Errorf("expected error")
		}

		if ran {
			t.Error("expected hub substep to not be run")
		}

		if st.Err() == nil {
			t.Error("expected error")
		}
	})

	t.Run("substep duration is printed", func(t *testing.T) {
		d := commanders.BufferStandardDescriptors(t)

		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, true)
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return nil
		})

		err = st.Complete("")
		if err != nil {
			d.Close()
			t.Errorf("unexpected err %#v", err)
		}

		stdout, stderr := d.Collect()
		d.Close()
		if len(stderr) != 0 {
			t.Errorf("unexpected stderr %#v", string(stderr))
		}

		actual := string(stdout)
		expected := fmt.Sprintf("\n%s took", idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG)
		// Use Contains since we don't know the actual step duration.
		if !strings.Contains(actual, expected) {
			t.Errorf("expected output %#v to end with %#v", actual, expected)
			t.Logf("actual: %s", actual)
			t.Logf("expected: %s", expected)
		}
	})

	t.Run("the step returns next actions when a substep fails", func(t *testing.T) {
		st, err := commanders.NewStep(idl.Step_INITIALIZE, &step.BufferedStreams{}, false)
		if err != nil {
			t.Errorf("unexpected err %#v", err)
		}

		expected := errors.New("oops")
		st.RunCLISubstep(idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, func(streams step.OutStreams) error {
			return expected
		})

		err = st.Complete("")
		var nextActionsErr cli.NextActions
		if !errors.As(err, &nextActionsErr) {
			t.Errorf("got %T, want %T", err, nextActionsErr)
		}
	})
}

func TestValidateStepWithStatusFile(t *testing.T) {
	dir := testutils.GetTempDir(t, "")
	defer testutils.MustRemoveAll(t, dir)

	resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", dir)
	defer resetEnv()

	path := filepath.Join(dir, "steps.json")
	testutils.MustWriteToFile(t, path, "{}")

	t.Run("fails when execute, finalize, revert ran before initialize has started", func(t *testing.T) {
		steps := []idl.Step{
			idl.Step_EXECUTE,
			idl.Step_FINALIZE,
			idl.Step_REVERT,
		}

		for _, s := range steps {
			err := commanders.ValidateStep(s)
			var validateErr commanders.ValidateStepError
			if !errors.As(err, &validateErr) {
				t.Fatalf("got %T, want %T", err, validateErr)
			}

			//TODO: check contents of error message
		}
	})

	t.Run("fails when execute, finalize ran before initialize has completed", func(t *testing.T) {
		testutils.MustWriteToFile(t, path, "{}")
		store := step.NewFileStore(path)
		err := store.Write(idl.Step_INITIALIZE, idl.Substep_INTERNAL_STEP_STATUS, idl.Status_RUNNING)
		if err != nil {
			t.Errorf("store.Write returned error %+v", err)
		}

		steps := []idl.Step{
			idl.Step_EXECUTE,
			idl.Step_FINALIZE,
		}

		for _, s := range steps {
			err := commanders.ValidateStep(s)
			var validateErr commanders.ValidateStepError
			if !errors.As(err, &validateErr) {
				t.Fatalf("%s: got %T, want %T", s, err, validateErr)
			}

			//TODO: check contents of error message
		}
	})

	t.Run("fails when initialize is run after execute has started", func(t *testing.T) {
		// ensure initialize has been run
		testutils.MustWriteToFile(t, path, "{}")
		store := step.NewFileStore(path)
		err := store.Write(idl.Step_EXECUTE, idl.Substep_INTERNAL_STEP_STATUS, idl.Status_RUNNING)
		if err != nil {
			t.Errorf("store.Write returned error %+v", err)
		}

		err = commanders.ValidateStep(idl.Step_INITIALIZE)
		var validateErr commanders.ValidateStepError
		if !errors.As(err, &validateErr) {
			t.Fatalf("got %T, want %T", err, validateErr)
		}

		//	expected := `To proceed with the upgrade, run "gpupgrade finalize".
		//To return the cluster to its original state, run "gpupgrade revert".`
		//	if nextActionsErr.NextActions != expected {
		//		t.Errorf("got %q want %q", nextActionsErr.NextActions, expected)
		//	}
	})

	t.Run("fails when finalize is run before execute", func(t *testing.T) {
		// ensure initialize has been run
		testutils.MustWriteToFile(t, path, "{}")
		store := step.NewFileStore(path)
		err := store.Write(idl.Step_INITIALIZE, idl.Substep_SAVING_SOURCE_CLUSTER_CONFIG, idl.Status_COMPLETE)
		if err != nil {
			t.Errorf("store.Write returned error %+v", err)
		}

		err = commanders.ValidateStep(idl.Step_FINALIZE)
		var validateErr commanders.ValidateStepError
		if !errors.As(err, &validateErr) {
			t.Fatalf("got %T, want %T", err, validateErr)
		}

		//		expected := `To proceed with the upgrade, run "gpupgrade execute".
		//To return the cluster to its original state, run "gpupgrade revert".`
		//		if nextActionsErr.NextActions != expected {
		//			t.Errorf("got %q want %q", nextActionsErr.NextActions, expected)
		//		}
		//	})
	})
}

func TestValidateStepWithNoStatusFile(t *testing.T) {
	resetEnv := testutils.SetEnv(t, "GPUPGRADE_HOME", "/does/not/exist")
	defer resetEnv()

	t.Run("does not error when initialize is run with no status file", func(t *testing.T) {
		err := commanders.ValidateStep(idl.Step_INITIALIZE)
		if err != nil {
			t.Errorf("ValidateStep returned error %+v", err)
		}
	})

	t.Run("fails with additional error context when execute, finalize, revert are run before initialize with no status file", func(t *testing.T) {
		steps := []idl.Step{
			idl.Step_EXECUTE,
			idl.Step_FINALIZE,
			idl.Step_REVERT,
		}

		for _, s := range steps {
			err := commanders.ValidateStep(s)
			var nextActionsErr cli.NextActions
			if !errors.As(err, &nextActionsErr) {
				t.Errorf("got %T, want %T", err, nextActionsErr)
			}
			//
			//			expected := `To begin the upgrade, run "gpupgrade initialize".`
			//			if nextActionsErr.NextActions != expected {
			//				t.Errorf("got %q want %q", nextActionsErr.NextActions, expected)
			//			}
			//
			//			expected = "status file: open /does/not/exist/status.json: no such file or directory"
			//			if !strings.HasPrefix(err.Error(), expected) {
			//				t.Errorf("want error %q to begin with %q", err, expected)
			//			}
		}
	})
}
