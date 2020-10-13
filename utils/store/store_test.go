//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package store

import "testing"

func TestMapStore(t *testing.T) {
	data := mapStore{}
	data.name = "/tmp/tryme.json"

	data.one = make(map[string]string)
	data.one["1"] = "a"
	data.one["2"] = "b"

	data.two = make(map[string]map[string]string)
	data.two["X"] = make(map[string]string)
	data.two["X"]["A"] = "22"
	data.two["X"]["B"] = "233"
	data.two["Y"] = make(map[string]string)
	data.two["Y"]["C"] = "44"
	data.two["Y"]["D"] = "55"

	err := data.Save()
	if err != nil {
		t.Fatalf("unexpected err: %#v", err)
	}

	readData := &mapStore{}
	err = data.Load(readData)
	if err != nil {
		t.Fatalf("unexpected err: %#v", err)
	}
}
