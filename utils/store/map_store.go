//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package store

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/xerrors"

	"github.com/greenplum-db/gpupgrade/utils"
)

type mapStore struct {
	name string
	one  map[string]string
	two  map[string]map[string]string
}

func (m *mapStore) Load(data *mapStore) error {
	file, err := os.Open(m.name)
	if err != nil {
		return xerrors.Errorf("opening file: %w", err)
	}
	defer file.Close()

	err = LoadJson(data, file)
	if err != nil {
		return xerrors.Errorf("reading file: %w", err)
	}

	return nil
}

// SaveConfig persists the hub's configuration to disk.
func (m *mapStore) Save() error {
	var buffer bytes.Buffer
	err := SaveJson(m, &buffer)
	if err != nil {
		return xerrors.Errorf("save config: %w", err)
	}
	fmt.Println("LENGTH:", buffer.Len())

	return utils.AtomicallyWrite(m.name, buffer.Bytes())
}
