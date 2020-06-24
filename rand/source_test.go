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

/* Test the rough uniformity of the crypto rand wrapper by using a Monte Carlo method [1] to compute pi [2].
 * The Monte Carlo method is sensitive to non-uniformity and in the event of a biased distribution we'd expect
 * to calculate a widely inaccurate value.
 *
 * [1] https://en.wikipedia.org/wiki/Monte_Carlo_method
 * [2] https://www.youtube.com/watch?v=AyBNnkYrSWY&t=149
 */
func TestCryptoRandomSource(t *testing.T) {
	rnd := rand.NewCryptoRand()

	circle := 0
	square := 0

	// Sample ten thousand random points within a bounding box.
	for i := 0; i < 10_000; i++ {
		x := rnd.Float64()
		y := rnd.Float64()

		// If the point falls within the bounds of a circle (eg. a radius from a origin point).
		// Then count that point as falling inside of the circle, as well as inside of the bounding box.
		if math.Pow(x, 2)+math.Pow(y, 2) <= 1.0 {
			circle++
		}
		square++
	}

	pi := 4.0 * float64(circle) / float64(square)
	assert.True(t, math.Abs(pi-math.Pi) < 0.1)
}
