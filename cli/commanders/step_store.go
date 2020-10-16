//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package commanders

import (
	"errors"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/cli"
	"github.com/greenplum-db/gpupgrade/idl"
	"github.com/greenplum-db/gpupgrade/step"
	"github.com/greenplum-db/gpupgrade/utils"
)

type StepStore struct {
	store *step.FileStore
}

func NewStepStore() (*StepStore, error) {
	path, err := utils.GetJSONFile(utils.GetStateDir(), StepsFileName)
	if err != nil {
		return &StepStore{}, xerrors.Errorf("getting %q file: %w", StepsFileName, err)
	}

	return &StepStore{
		store: step.NewFileStore(path),
	}, nil
}

func (s *StepStore) Write(stepName idl.Step, status idl.Status) error {
	err := s.store.Write(stepName, idl.Substep_STEP_STATUS, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepStore) Read(stepName idl.Step) (idl.Status, error) {
	status, err := s.store.Read(stepName, idl.Substep_STEP_STATUS)
	if err != nil {
		return idl.Status_UNKNOWN_STATUS, err
	}

	return status, nil
}

func (s *StepStore) HasStepStarted(step idl.Step) (bool, error) {
	return s.HasStatus(step, func(status idl.Status) bool {
		return status != idl.Status_UNKNOWN_STATUS
	})
}

func (s *StepStore) HasStepCompleted(step idl.Step) (bool, error) {
	return s.HasStatus(step, func(status idl.Status) bool {
		return status == idl.Status_COMPLETE
	})
}

func (s *StepStore) HasStatus(step idl.Step, check func(status idl.Status) bool) (bool, error) {
	status, err := s.Read(step)
	if err != nil {
		return false, err
	}

	return check(status), nil
}

type validStep struct {
	idl.Step
	nextAction string
}

type stepConditions struct {
	notStarted []validStep
	completed  []validStep
}

const NextActionRunInitialize = `To begin the upgrade, run "gpupgrade initialize".`

const NextActionRunExecute = `To proceed with the upgrade, run "gpupgrade execute".
To return the cluster to its original state, run "gpupgrade revert".`

const NextActionRunFinalize = `To proceed with the upgrade, run "gpupgrade finalize".
To return the cluster to its original state, run "gpupgrade revert".`

const NextActionCompleteFinalize = `To proceed with the upgrade, run "gpupgrade finalize.`

// stepConditions are conditions are are expected to have been met for the
// current step. The next action message is printed if the condition is not met.
var validate = map[idl.Step]stepConditions{
	idl.Step_INITIALIZE: {
		notStarted: []validStep{
			{idl.Step_EXECUTE, NextActionRunExecute},
			{idl.Step_FINALIZE, NextActionRunFinalize},
		},
	},
	idl.Step_EXECUTE: {
		notStarted: []validStep{
			{idl.Step_FINALIZE, NextActionRunFinalize},
		},
		completed: []validStep{
			{idl.Step_INITIALIZE, NextActionRunInitialize},
		},
	},
	idl.Step_FINALIZE: {
		completed: []validStep{
			{idl.Step_INITIALIZE, NextActionRunInitialize},
			{idl.Step_EXECUTE, NextActionRunExecute},
		},
	},
	idl.Step_REVERT: {
		notStarted: []validStep{
			{idl.Step_FINALIZE, NextActionCompleteFinalize},
		},
		completed: []validStep{
			{idl.Step_INITIALIZE, NextActionRunInitialize},
		},
	},
}

func (s *StepStore) ValidateStep(currentStep idl.Step) (err error) {
	stepErr := errors.New(`gpupgrade commands must be issued in correct order
  1. initialize   runs pre-upgrade checks and prepares the cluster for upgrade
  2. execute      upgrades the master and primary segments to the target
                  Greenplum version
  3. finalize     upgrades the standby master and mirror segments to the target
                  Greenplum version. Revert cannot be run after finalize has started.
Use "gpupgrade --help" for more information`)

	conditions := validate[currentStep]

	// ensure specified steps have not started
	for _, st := range conditions.notStarted {
		started, err := s.HasStepStarted(st.Step)
		if err != nil {
			return err
		}

		if started {
			return cli.NewNextActions(stepErr, currentStep.String(), false, st.nextAction)
		}
	}

	// check if required steps have completed
	for _, st := range conditions.completed {
		completed, err := s.HasStepCompleted(st.Step)
		if err != nil {
			return err
		}

		if !completed {
			return cli.NewNextActions(stepErr, currentStep.String(), false, st.nextAction)
		}
	}

	return nil
}
