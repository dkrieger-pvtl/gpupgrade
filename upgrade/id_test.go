// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package upgrade_test

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/greenplum-db/gpupgrade/upgrade"
)

func TestID(t *testing.T) {
	t.Run("NewID gives a unique identifier for each run", func(t *testing.T) {
		one := upgrade.NewID()
		two := upgrade.NewID()

		if one == two {
			t.Errorf("second generated ID was equal to first ID (%d)", one)
		}
	})

	t.Run("String gives a base64 representation of the ID", func(t *testing.T) {
		var id upgrade.ID

		expected := "AAAAAAAAAAA" // all zeroes in base64. 8 bytes decoded -> 11 bytes encoded
		if id.String() != expected {
			t.Errorf("String() returned %q, want %q", id.String(), expected)
		}
	})
}

// TestNewIDCrossProcess ensures that NewID returns different results across
// invocations of an executable (i.e. that the ID source is seeded correctly).
func TestNewIDCrossProcess(t *testing.T) {
	cmd1 := idCommand()
	cmd2 := idCommand()

	out1, err := cmd1.Output()
	if err != nil {
		t.Errorf("first execution: unexpected error %+v", err)
	}

	out2, err := cmd2.Output()
	if err != nil {
		t.Errorf("second execution: unexpected error %+v", err)
	}

	if string(out1) == string(out2) {
		t.Errorf("second generated ID was equal to first ID (%s)", string(out1))
	}
}

// idCommand creates an exec.Cmd that will run upgrade.NewID() in a brand-new
// process. It uses the TestIDCommand entry point to do its work.
func idCommand() *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestIDCommand")
	cmd.Env = append(cmd.Env, "GO_RUN_NEW_ID=1")
	return cmd
}

// TestIDCommand is the entry point for the idCommand(). It simply prints the
// result of an upgrade.NewID().
func TestIDCommand(_ *testing.T) {
	if os.Getenv("GO_RUN_NEW_ID") != "1" {
		return
	}

	fmt.Printf("%d", upgrade.NewID())
	os.Exit(0)
}

// Make sure we are filtering out "--".  Empirical testing shows about 300 iterations
//  are required to hit a "--", so we choose 10000 to ensure we'd catch an erroneous
//  implementation.  This test takes 10ms on my laptop.
func TestNoDoubleDash(t *testing.T) {
	for i := 0; i < 10000; i++ {
		id := upgrade.NewID()
		if strings.Contains(id.String(), "--") {
			t.Fatalf("id %s contains --", id)
		}
	}
}

func TestNoDoubleDash2(t *testing.T) {
	t.Run("returns an ID with no --", func(t *testing.T) {
		mustNotContain(t, "--", upgrade.NewID())
	})

	t.Run("explicitly hits no -- check", func(t *testing.T) {
		// return an ID containing "--" on the first call, and one that
		// does not contain "--" on the second call
		called := false
		d := upgrade.SetRandomBytes(func(b []byte) (n int, err error) {
			if !called {
				called = true

				if len(b) != 8 {
					t.Errorf("only support an 8 byte buffer")
				}

				// in base64, '-' is 62, so make the first two characters '--' by having
				// the bit pattern be "0b111110111110...' (https://tools.ietf.org/html/rfc4648)
				b[0] = 62<<2 + 3 // 0b11111011
				b[1] = 14 << 4   //         0b11100000
				for i := 2; i < 8; i++ {
					b[i] = 0 // 0b000000 is 'A' in filesystem safe base64
				}
				mustContain(t, "--", upgrade.ID(binary.LittleEndian.Uint64(b)))

				return 8, nil
			} else {
				for i := 0; i < 8; i++ {
					b[i] = 0 // 0b000000 is 'A' in filesystem safe base64
				}
				mustNotContain(t, "--", upgrade.ID(binary.LittleEndian.Uint64(b)))
				return 8, nil
			}
		})
		defer d()

		mustNotContain(t, "--", upgrade.NewID())
	})

	t.Run("panics if random byte generation returns an error", func(t *testing.T) {
		d := upgrade.SetRandomBytes(func(b []byte) (n int, err error) {
			return 0, errors.New("intentional panic")
		})
		defer d()

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("did not panic")
			}
		}()

		upgrade.NewID()
	})
}

func mustNotContain(t *testing.T, s string, id upgrade.ID) {
	if strings.Contains(id.String(), s) {
		t.Errorf("expected no -- in ID, got %s", id.String())
	}
}
func mustContain(t *testing.T, s string, id upgrade.ID) {
	if !strings.Contains(id.String(), s) {
		t.Errorf("expected no -- in ID, got %s", id.String())
	}
}
