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

// Package password provides methods for generating secure random passwords,
// and shared constants such as common character sets.
package password

import (
	"errors"

	"github.com/kloeckner-i/can-haz-password/rand"
)

// Characters for use as character classes. Exported for utility purposes.
const (
	DigitCharacters     = "0123456789"
	UppercaseCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LowercaseCharacters = "abcdefghijklmnopqrstuvwxyz"
	// OWASP recommended password special characters.
	SpecialCharacters        = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	URLSafeSpecialCharacters = "-_"
)

// ErrInvalidPasswordRejection is returned when a password rule rejects an excessive number of passwords.
var ErrInvalidPasswordRejection = errors.New("password rule rejected too many passwords")

// Configuration sets the properties of the generated password.
type Configuration struct {
	/* Minimum length of the password.
	 * The actual length will be random, and in the range minimum_length <= length <= 1.5 * minimum_length.
	 * Random lengths allow for minimum complexity requirements to be met while not enforcing a strict
	 * composition (eg. exactly 2 digits, and exactly 1 special character).
	 */
	Length int
	// CharacterClasses configuration, eg. the list of characters and the minimum quantity to include.
	CharacterClasses []CharacterClassConfiguration
}

// CharacterClassConfiguration configures the character composition of the password.
type CharacterClassConfiguration struct {
	// The Characters included in this character class.
	Characters string
	// The Minimum number of characters from this character class to include in the password.
	Minimum int
}

// Rule is used to set the behavior of the random password generator.
type Rule interface {
	// Return the configuration associated with this rule.
	Config() *Configuration
	// Is the password valid?
	Valid(password []rune) bool
}

// Generator generates random passwords matching a rule.
type Generator struct {
	characterSource rand.WeightedRandomSet
	passwordRule    Rule
}

// NewGenerator constructs a random password generator from a rule.
func NewGenerator(passwordRule Rule) Generator {
	return Generator{
		characterSource: buildCharacterSource(passwordRule.Config()),
		passwordRule:    passwordRule,
	}
}

// Generate a new random password.
func (g Generator) Generate() (string, error) {
	// Prevent the possibility of ending up in an infinite loop due to a bad rule.
	const maxInvalidPasswordRejections = 10

	config := g.passwordRule.Config()
	password := make([]rune, 0)

	for invalidPasswordRejections := 0; invalidPasswordRejections < maxInvalidPasswordRejections; {
		if g.complete(config, password) {
			return string(password), nil
		}

		// Append the next character to the password.
		password = append(password, g.characterSource.Next())

		// Reject the creation of passwords that the rule would consider "invalid".
		if !g.passwordRule.Valid(password) {
			// Rollback the last character.
			password = password[:len(password)-1]
			invalidPasswordRejections++

			continue
		}

		// If we exceed the maximum length, just try again from the start.
		// This sets an upper bound on the long tail of the distribution.
		if len(password) > int(float64(config.Length)*1.5) {
			password = make([]rune, 0)
		}
	}

	return "", ErrInvalidPasswordRejection
}

// Have we completed the password? Eg. have we met the minimum length requirement and all of the complexity
// requirements?
func (g Generator) complete(config *Configuration, password []rune) bool {
	if len(password) < config.Length {
		return false
	}

	for _, characterClass := range config.CharacterClasses {
		if occurrencesOfCharacters(password, []rune(characterClass.Characters)) < characterClass.Minimum {
			return false
		}
	}

	return true
}

// The character source is backed by a weighted random set that returns values according to a distribution
// consistent with the desired composition of the final password.
func buildCharacterSource(config *Configuration) rand.WeightedRandomSet {
	// The total number of characters in all of our character classes.
	totalCount := 0
	for _, characterClass := range config.CharacterClasses {
		totalCount += characterClass.Minimum
	}

	entries := make([]rand.WeightedRandomEntry, 0)

	// Build weighted entries for each character class.
	for _, characterClass := range config.CharacterClasses {
		// The probability of a character belonging to this particular character class.
		// This is tuned via the minimum complexity requirement, and sets the composition of the generated password.
		probability := float64(characterClass.Minimum) / float64(totalCount)
		// Add an entry for every character in the character class. Each character in the class is equally weighted.
		entries = addCharactersToWeightedRandomSet(entries, []rune(characterClass.Characters), probability)
	}

	return rand.NewWeightedRandomSet(entries)
}

// If the character class is specified, calculate the desired probability of each character in the character class,
// and add them to the weighted random set.
func addCharactersToWeightedRandomSet(
	entries []rand.WeightedRandomEntry, characterClass []rune, probability float64) []rand.WeightedRandomEntry {
	// If the rule supplied a character class with a zero probability of being selected for some bizarre reason.
	// Lets avoid creating an entry for it.
	if probability > 0.0 {
		for _, v := range characterClass {
			entries = append(entries, rand.WeightedRandomEntry{
				Character: v,
				// The probability of each individual character being selected is equal to the probability of this
				// class of characters being selected, divided by the total number of characters in this class.
				Weight: probability / float64(len(characterClass)),
			})
		}
	}

	return entries
}

// Count the number of occurrences of the character class in a password.
// Used for determining if a generated password meets the complexity rules.
func occurrencesOfCharacters(password, characterClass []rune) int {
	freq := make(map[rune]int)
	total := 0

	// Sum up the total number of occurrences of each character in the password.
	for _, c := range password {
		freq[c]++
	}

	// Go through all the characters contained in the character class and count the total number of occurrences.
	for _, c := range characterClass {
		if count, ok := freq[c]; ok {
			total += count
		}
	}

	return total
}
