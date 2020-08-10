//  Copyright (c) 2017-2020 VMware, Inc. or its affiliates
//  SPDX-License-Identifier: Apache-2.0

package gpupgrade

import "math"

func sieve(max int) []int {

	notprimes := make([]bool, max+1)
	notprimes[0], notprimes[1] = true, true

	limit := int(math.Sqrt(float64(max)))
	for i := 2; i <= limit; i++ {
		if !notprimes[i] {
			for j := 2 * i; j <= max; j += i {
				notprimes[j] = true
			}
		}
	}

	var primes []int
	for prime, notPrime := range notprimes {
		if !notPrime {
			primes = append(primes, prime)
		}
	}

	return primes
}
