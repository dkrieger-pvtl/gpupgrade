package hub_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"

	"github.com/greenplum-db/gpupgrade/hub"
)

type agentStarterMock struct {
	called   map[string]bool
	stateDir string
}

func (a *agentStarterMock) StartAgent(hostname, stateDir string) (err error) {
	fmt.Printf("mock starting agent on host %s using stateDir %s", hostname, stateDir)
	a.called[hostname] = true
	a.stateDir = stateDir
	return nil
}

type agentStarterMockError struct {
	called map[string]bool
}

func (a *agentStarterMockError) StartAgent(hostname, stateDir string) error {
	a.called[hostname] = true
	if hostname == "a" || hostname == "b" {
		return errors.New(fmt.Sprintf("%s", hostname))
	}
	return nil
}

func TestStartAgent(t *testing.T) {

	t.Run("basic logical test", func(t *testing.T) {
		stateDir := "NewYork"
		called := make(map[string]bool)
		asm := agentStarterMock{called, ""}
		hostnames := []string{"hostname1", "hostname2"}

		err := hub.StartAgentsSubStep(hostnames, stateDir, &asm)
		if err != nil {
			t.Errorf("got unexpected err: %#v", err)
		}
		if !asm.called["hostname1"] {
			t.Errorf("expected StartAgent to be called on hostname1")
		}
		if !asm.called["hostname2"] {
			t.Errorf("expected StartAgent to be called on hostname2")
		}
		if asm.stateDir != stateDir {
			t.Errorf("expected StartAgent to use the stateDir %s passed to it, but got %s",
				stateDir, asm.stateDir)
		}
	})

	t.Run("returns an error for each hostname that errors", func(t *testing.T) {
		called := make(map[string]bool)
		asm := agentStarterMockError{called}
		err := hub.StartAgentsSubStep([]string{"a", "b", "c"}, "x", &asm)
		if err == nil {
			t.Errorf("expected StartAgent to error, got nil")
		}

		for _, hostname := range []string{"a", "b", "c"} {
			if !asm.called[hostname] {
				t.Errorf("expected StartAgent to be called on hostname %s", hostname)
			}
		}
		if merr, ok := err.(*multierror.Error); ok {
			a := false
			b := false
			c := false
			for _, serr := range merr.Errors {
				if serr.Error() == "a" {
					a = true
				}
				if serr.Error() == "b" {
					b = true
				}
				if serr.Error() == "c" {
					c = true
				}
			}
			if !a {
				t.Errorf("expected error on hostname a, found none")
			}
			if !b {
				t.Errorf("expected error on hostname b, found none")
			}
			if c {
				t.Errorf("expected no error on hostname c, found one")
			}
		}

	})

}
