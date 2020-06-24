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

package rand

import (
	"math/rand"

	"github.com/geozelot/intree"
)

// WeightedRandomSet is a weighted random set that can be used to select elements in a non-uniformally distributed
// fashion. This will typically be weighted to match the desired composition of a generated password.
//
// The algorithm and data structure used here is based around an interval tree [1] of accumulated weights.
// We take an unordered collection of entries, and for each entry we insert it into the interval tree in the interval
// total <= x < total + entry_weight where total is the sum of all of the previous weights already stored in the tree.
// The insertion order is not important. The end result of this is an interval tree containing a sequence of immediately
// adjacent intervals of varying width / weight. By generating a random value in the range 0 <= x < total we guarantee
// an interval exists in the tree which contains the value. The probability that the random value will fall in any
// individual interval is equal to the ratio of its width to the total width of the tree. This is the mechanism that
// allows us to weight/bias the resulting random distribution.
//
// [1] https://en.wikipedia.org/wiki/Interval_tree
type WeightedRandomSet struct {
	randSource *rand.Rand
	// An interval tree used internally in the weighting algorithm.
	tree *intree.INTree
	// The character associated with each element in the interval tree.
	values []rune
	// The total weight of the tree (eg. the sum of all the entry weights).
	totalWeight float64
}

// WeightedRandomEntry is an entry in the weighted random set.
type WeightedRandomEntry struct {
	Value  rune
	Weight float64
}

// NewWeightedRandomSet is used to construct a new weighted random set from a collection of entries.
func NewWeightedRandomSet(entries []WeightedRandomEntry) WeightedRandomSet {
	totalWeight := 0.0
	ranges := make([]intree.Bounds, len(entries))
	values := make([]rune, len(entries))

	for i, w := range entries {
		ranges[i] = &interval{min: totalWeight, max: totalWeight + w.Weight}
		values[i] = w.Value
		totalWeight += w.Weight
	}

	return WeightedRandomSet{
		randSource:  NewCryptoRand(),
		tree:        intree.NewINTree(ranges),
		values:      values,
		totalWeight: totalWeight,
	}
}

// Next returns the next value in the weighted random sequence.
func (rs WeightedRandomSet) Next() rune {
	// A single entry in the interval tree is guaranteed to exist for the value.
	// So we can easily query the tree and return the first value without checking.
	return rs.values[rs.tree.Including(rs.randSource.Float64() * rs.totalWeight)[0]]
}

// A simple floating point interval implementation for the interval tree.
type interval struct {
	min, max float64
}

func (i *interval) Limits() (float64, float64) {
	return i.min, i.max
}
