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

package rand_test

import (
	"math"
	"testing"

	"github.com/kloeckner-i/can-haz-password/rand"
	"github.com/stretchr/testify/assert"
)

// Ensure the weighted random returns a biased stream of characters.
func TestWeightedRandomSet(t *testing.T) {
	// A collection of random entries with a triangle like distribution of probability.
	// Weight being akin to the probability of being selected.
	entries := []rand.WeightedRandomEntry{
		{Character: 0, Weight: 1.0},
		{Character: 1, Weight: 2.0},
		{Character: 2, Weight: 3.0},
		{Character: 3, Weight: 4.0},
		{Character: 4, Weight: 5.0},
		{Character: 5, Weight: 4.0},
		{Character: 6, Weight: 3.0},
		{Character: 7, Weight: 2.0},
		{Character: 8, Weight: 1.0},
	}

	// Construct a new weighted random set.
	rnd := rand.NewWeightedRandomSet(entries)

	// Read a large number of values from the weighted random set.
	const samples = 10_000
	values := make([]rune, samples)

	for i := 0; i < samples; i++ {
		values[i] = rnd.Next()
	}

	// Compute the percentage of occurrences for each value.
	probabilities := make([]float64, 9)
	for _, v := range values {
		probabilities[v] += 1.0
	}

	for v := range probabilities {
		probabilities[v] /= float64(len(values))
	}

	// A set of expected probability of occurring, based on the weights of our entries.
	// The index of the array being the value of the entry. Should follow a triangle like distribution.
	expectedProbabilities := []float64{
		0.04,
		0.08,
		0.12,
		0.16,
		0.2,
		0.16,
		0.12,
		0.08,
		0.04,
	}

	assert.True(t, approximatelyEqual(expectedProbabilities, probabilities, 0.02))
}

// Compare an array of floating point values for approximate equality, epsilon being an upper bound on the error.
func approximatelyEqual(expected, actual []float64, epsilon float64) bool {
	for i, v := range expected {
		if math.Abs(v-actual[i]) > epsilon {
			return false
		}
	}

	// Only true if all elements are within the error bound.
	return true
}
