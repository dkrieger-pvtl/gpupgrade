// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package rsync_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/testutils"
	"github.com/greenplum-db/gpupgrade/utils/rsync"
)

func writeToFile(filepath string, contents []byte, t *testing.T) {
	err := ioutil.WriteFile(filepath, contents, 0644)

	if err != nil {
		t.Fatalf("error writing file '%v'", filepath)
	}
}

func TestRsync(t *testing.T) {
	emptyHost := ""

	testhelper.SetupTestLogger()

	if _, err := exec.LookPath("rsync"); err != nil {
		t.Skipf("tests require rsync (%v)", err)
	}

	// These are "live" integration tests. Plug exec.Command back into the
	// system.
	rsync.SetRsyncCommand(exec.Command)
	defer func() { rsync.SetRsyncCommand(nil) }()

	t.Run("when createSubdirs is false, it copies data from a source directory to the top level of the target directory", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		filename := "this_is_my_file_name.txt"
		writeToFile(filepath.Join(sourceDir, filename), []byte("hi"), t)

		if err := rsync.RsyncWithoutStream(sourceDir, "localhost", targetDir, []string{"--archive", "--delete"}, []string{}); err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		targetContents, _ := ioutil.ReadFile(filepath.Join(targetDir, "/", filename))

		if !bytes.Equal(targetContents, []byte("hi")) {
			t.Errorf("target directory file 'hi' contained %v, wanted %v",
				targetContents,
				"hi")
		}
	})

	t.Run("when createSubdirs is true, it copies data from a source directory to a subdir of the target directory", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		filename := "this_is_my_file_name.txt"
		writeToFile(filepath.Join(sourceDir, filename), []byte("hi"), t)

		if err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{}, step.DevNullStream, true); err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		basedir := filepath.Base(sourceDir)
		targetContents, _ := ioutil.ReadFile(filepath.Join(targetDir, "/", basedir, filename))

		if !bytes.Equal(targetContents, []byte("hi")) {
			t.Errorf("target directory file 'hi' contained %v, wanted %v",
				targetContents,
				"hi")
		}
	})

	t.Run("it copies multiple source directories to the target directory", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		sourceDir2 := testutils.GetTempDir(t, "rsync-source2")
		defer testutils.MustRemoveAll(t, sourceDir2)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		filename := "this_is_my_file_name.txt"
		writeToFile(filepath.Join(sourceDir, filename), []byte("hi"), t)

		filename2 := "this_is_my_file_name_2.txt"
		writeToFile(filepath.Join(sourceDir2, filename2), []byte("hi_2"), t)

		if err := rsync.Rsync([]string{sourceDir, sourceDir2}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{}, step.DevNullStream, true); err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		basedir := filepath.Base(sourceDir)
		targetContents, _ := ioutil.ReadFile(filepath.Join(targetDir, "/", basedir, filename))

		if !bytes.Equal(targetContents, []byte("hi")) {
			t.Errorf("target directory file 'hi' contained %v, wanted %v",
				targetContents,
				"hi")
		}

		basedir2 := filepath.Base(sourceDir2)
		targetContents2, _ := ioutil.ReadFile(filepath.Join(targetDir, "/", basedir2, filename2))

		if !bytes.Equal(targetContents2, []byte("hi_2")) {
			t.Errorf("target directory file 'hi_2' contained %v, wanted %v",
				targetContents2,
				"hi_2")
		}
	})

	t.Run("a passed in stream contains the copied filename in verbose mode", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		filename := "this_is_my_file_name.txt"
		writeToFile(filepath.Join(sourceDir, filename), []byte("hi"), t)

		streams := &step.BufferedStreams{}
		if err := rsync.RsyncWithStream(sourceDir, emptyHost, targetDir, []string{"--archive", "--verbose"}, []string{}, streams); err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}
		if !strings.Contains(streams.StdoutBuf.String(), filename) {
			t.Errorf("expected stdout to contain filename: %s", streams.StdoutBuf.String())
		}

		targetContents, _ := ioutil.ReadFile(filepath.Join(targetDir, "/", filename))

		if !bytes.Equal(targetContents, []byte("hi")) {
			t.Errorf("target directory file 'hi' contained %v, wanted %v",
				targetContents,
				"hi")
		}
	})

	t.Run("it removes files that existed in the target directory before the sync", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		writeToFile(filepath.Join(targetDir, "target-file-that-should-get-removed"), []byte("goodbye"), t)

		if err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{}, step.DevNullStream, false); err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		_, statError := os.Stat(filepath.Join(targetDir, "target-file-that-should-get-removed"))

		if os.IsExist(statError) {
			t.Errorf("target directory file 'target-file-that-should-get-removed' should not exist, but it does")
		}
	})

	t.Run("it does not copy files from the source directory when in the exclusion list", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		writeToFile(filepath.Join(sourceDir, "source-file-that-should-get-excluded"), []byte("goodbye"), t)

		err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{"source-file-that-should-get-excluded"}, step.DevNullStream, false)
		if err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		_, statError := os.Stat(filepath.Join(targetDir, "source-file-that-should-get-excluded"))

		if os.IsExist(statError) {
			t.Errorf("target directory file 'source-file-that-should-get-excluded' should not exist, but it does")
		}
	})

	t.Run("it preserves files in the target directory when in the exclusion list", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := testutils.GetTempDir(t, "rsync-target")
		defer testutils.MustRemoveAll(t, targetDir)

		writeToFile(filepath.Join(sourceDir, "source-file-that-should-get-copied"), []byte("new file"), t)
		writeToFile(filepath.Join(targetDir, "target-file-that-should-get-ignored"), []byte("i'm still here"), t)
		writeToFile(filepath.Join(targetDir, "another-target-file-that-should-get-ignored"), []byte("i'm still here"), t)

		err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{"target-file-that-should-get-ignored", "another-target-file-that-should-get-ignored"}, step.DevNullStream, false)
		if err != nil {
			t.Errorf("Rsync() returned error %+v", err)
		}

		_, statError := os.Stat(filepath.Join(targetDir, "target-file-that-should-get-ignored"))

		if os.IsNotExist(statError) {
			t.Error("target directory file 'target-file-that-should-get-ignored' should still exist, but it does not")
		}

		_, statError = os.Stat(filepath.Join(targetDir, "another-target-file-that-should-get-ignored"))

		if os.IsNotExist(statError) {
			t.Error("target directory file 'another-target-file-that-should-get-ignored' should still exist, but it does not")
		}

		_, statError = os.Stat(filepath.Join(targetDir, "source-file-that-should-get-copied"))

		if os.IsNotExist(statError) {
			t.Error("target directory file 'source-file-that-should-get-copied' should exist, but does not")
		}
	})

	t.Run("it bubbles up exec.ExitError errors as rsync errors", func(t *testing.T) {
		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := "/tmp/some/invalid/target/dir"
		defer testutils.MustRemoveAll(t, targetDir)

		writeToFile(filepath.Join(sourceDir, "some-file"), []byte("hi"), t)

		err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{""}, step.DevNullStream, false)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		//var rsyncError rsync.RsyncError
		//
		//if !xerrors.As(err, &rsyncError) {
		//	t.Errorf("got error %#v, wanted type %T", err, rsyncError)
		//}
		//
		//expected := "rsync: mkdir \"/tmp/some/invalid/target/dir\" failed"
		//if !strings.Contains(rsyncError.Error(), expected) {
		//	t.Errorf("got %v, expected substring %s",
		//		err.Error(), expected)
		//}
	})

	t.Run("it bubbles up exec.Error errors as rsync errors", func(t *testing.T) {
		originalPath := destroyPath()
		defer restorePath(originalPath)

		sourceDir := testutils.GetTempDir(t, "rsync-source")
		defer testutils.MustRemoveAll(t, sourceDir)

		targetDir := "/tmp/some/invalid/target/dir"
		defer testutils.MustRemoveAll(t, targetDir)

		writeToFile(filepath.Join(sourceDir, "some-file"), []byte("hi"), t)

		err := rsync.Rsync([]string{sourceDir}, emptyHost, targetDir, []string{"--archive", "--delete"}, []string{""}, step.DevNullStream, false)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		//var rsyncError rsync.RsyncError
		//
		//if !xerrors.As(err, &rsyncError) {
		//	t.Errorf("got error %#v, wanted type %T", err, rsyncError)
		//}
		//
		//expected := "exec: \"rsync\": executable file not found in $PATH"
		//if !strings.Contains(rsyncError.Error(), expected) {
		//	t.Errorf("got %v, wanted %s",
		//		err.Error(), "exec: \"rsync\": executable file not found in $PATH")
		//}
	})
}

func restorePath(originalPath string) {
	os.Setenv("PATH", originalPath)
}

func destroyPath() string {
	var originalPath = os.Getenv("PATH")

	os.Setenv("PATH", "/nothing")

	return originalPath
}
