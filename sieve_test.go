//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package gpupgrade

import (
	"reflect"
	"testing"
)

func TestSieve(t *testing.T) {
	r := sieve(10)
	expected := []int{2, 3, 5, 7}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("got %v, expected %v", r, expected)
	}
}

func TestSieve2(t *testing.T) {
	r := sieve(1000000)
	expected := 16433
	for _, val := range r {
		if val == expected {
			return
		}
	}
	t.Errorf("got %v, did not contain expected %d", r, expected)
}
