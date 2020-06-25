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

package password_test

import (
	"fmt"
	"math"
	"regexp"
	"testing"

	"github.com/kloeckner-i/can-haz-password/password"
	"github.com/stretchr/testify/assert"
)

func TestPasswordGenerator(t *testing.T) {
	generator := password.NewGenerator(newDummyPasswordRule())

	passwords := make([]string, 0)

	// Generate ten thousand short passwords.
	for i := 0; i < 10_000; i++ {
		password, err := generator.Generate()
		assert.Nil(t, err)

		passwords = append(passwords, password)
	}

	// Compute statistics on the length of the generated passwords.
	minLength := 12
	maxLength := 0
	meanLength := 0.0

	for _, password := range passwords {
		if len(password) < minLength {
			minLength = len(password)
		}

		if len(password) > maxLength {
			maxLength = len(password)
		}

		meanLength += float64(len(password))
	}

	meanLength /= float64(len(passwords))

	// Ensure the password length is bounded and the average length falls within the expected range.
	assert.Equal(t, 8, minLength)
	assert.Equal(t, 12, maxLength)
	assert.True(t, meanLength > 9.0 && meanLength < 10.0)

	// Is the distribution of characters as expected?
	// Calculate this by summing up the total number of occurrences for each character.
	total := 0
	counts := make(map[rune]float64)

	for _, password := range passwords {
		for _, v := range password {
			counts[v] += 1.0
			total++
		}
	}

	for i, count := range counts {
		counts[i] = count / float64(total)
	}

	allLetters := append([]rune(password.UppercaseCharacters), []rune(password.LowercaseCharacters)...)
	// From the password rule we expect 3/8s of all characters to be a letter of either case.
	expectedLetterProbability := 3.0 / (8.0 * float64(len(allLetters)))

	for _, c := range allLetters {
		assert.True(t, math.Abs(counts[c]-expectedLetterProbability) < 0.01)
	}

	digits := []rune(password.DigitCharacters)
	// From the password rule we expect 3/8s of all characters to be a digit.
	expectedDigitProbability := 3.0 / (8.0 * float64(len(digits)))

	for _, c := range digits {
		assert.True(t, math.Abs(counts[c]-expectedDigitProbability) < 0.01)
	}

	specialCharacters := []rune(password.URLSafeSpecialCharacters)
	// From the password rule we expect 1/8 of all characters to be a special character.
	expectedSpecialCharacterProbability := 1.0 / (8.0 * float64(len(specialCharacters)))

	for _, c := range specialCharacters {
		fmt.Println(counts[c], expectedSpecialCharacterProbability)
		// We allow a slightly larger error margin for special characters, as the combination of low normal
		// prevalence, combined with the minimum complexity rule, leads to them being slightly overrepresented.
		// To the order of 30 - 40% but this is an expected consequence of the minimum complexity rules.
		assert.True(t, math.Abs(counts[c]-expectedSpecialCharacterProbability) < 0.025)
	}

	// Do the passwords contain any invalid characters? Eg. characters outside of the expected classes.
	valid := regexp.MustCompile(`^[a-zA-Z0-9_-]{8,12}$`)
	// Are any of the passwords "invalid" somehow?
	invalid := regexp.MustCompile(`[-]{2,}`)

	for _, password := range passwords {
		assert.Truef(t, valid.MatchString(password), "password '%v' was not valid", password)
		assert.Falsef(t, invalid.MatchString(password), "password '%v' contains invalid double dash sequence", password)
	}
}

// Test that a broken password rule (eg. one that rejects every password) returns an error.
func TestPasswordGeneratorReturnsErrorForBrokenRule(t *testing.T) {
	generator := password.NewGenerator(newBrokenPasswordRule())

	_, err := generator.Generate()

	assert.Equal(t, password.ErrInvalidPasswordRejection, err)
}

// Short passwords with hybris style invalid characters.
type dummyPasswordRule struct {
	invalid *regexp.Regexp
}

func newDummyPasswordRule() *dummyPasswordRule {
	return &dummyPasswordRule{
		// Does not support consecutive dashes.
		invalid: regexp.MustCompile(`[-]{2,}`),
	}
}

func (r *dummyPasswordRule) Config() *password.Configuration {
	return &password.Configuration{
		Length: 8,
		CharacterClasses: []password.CharacterClassConfiguration{
			{Characters: password.LowercaseCharacters + password.UppercaseCharacters, Minimum: 3},
			{Characters: password.DigitCharacters, Minimum: 3},
			{Characters: password.URLSafeSpecialCharacters, Minimum: 1},
		},
	}
}

func (r *dummyPasswordRule) Valid(password []rune) bool {
	return !r.invalid.MatchString(string(password))
}

// A password rule that always rejects the proposed password.
type brokenPasswordRule struct{}

func newBrokenPasswordRule() *brokenPasswordRule {
	return &brokenPasswordRule{}
}

func (r *brokenPasswordRule) Config() *password.Configuration {
	return &password.Configuration{
		Length: 8,
		CharacterClasses: []password.CharacterClassConfiguration{
			{Characters: password.LowercaseCharacters + password.UppercaseCharacters, Minimum: 8},
		},
	}
}

func (r *brokenPasswordRule) Valid(password []rune) bool {
	// Always rejects the password.
	return false
}
