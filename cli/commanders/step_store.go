//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package commanders

import (
	"golang.org/x/xerrors"

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
