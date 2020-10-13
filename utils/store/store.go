//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package store

import (
	"encoding/json"
	"io"
)

type Store interface {
	Load(t interface{}) error
	Save() error
}

func LoadJson(t interface{}, r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(t)
}

func SaveJson(t interface{}, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(t)
}
