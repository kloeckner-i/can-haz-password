/*
 * Copyright 2020 kloeckner.i GmbH
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package rand provides primitives and utilities for generating and working with secure
// random numbers. Including generating non-uniform random number distributions.
package rand

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

// NewCryptoRand is used to construct a new `math/rand` instance that is backed by `crypto/rand`.
// This allows the user to securely generate random values, and use the utility methods provided by `math/rand`.
func NewCryptoRand() *mrand.Rand {
	return mrand.New(new(cryptoRandomSource))
}

// cryptoRandomSource is a `math/rand` compatible source for providing secure random numbers.
// It operates by implementing the `rand.Source` interface for the operating systems cryptographic random source.
type cryptoRandomSource struct{}

// Seeding is explicitly ignored, as the operating system will seed its cryptographic source for us.
func (r cryptoRandomSource) Seed(_ int64) {}

func (r cryptoRandomSource) Int63() int64 {
	return int64(r.Uint64() & ^uint64(1<<63))
}

// Read eight bytes from the operating systems cryptographic random source.
func (r cryptoRandomSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		panic(err)
	}

	return v
}
