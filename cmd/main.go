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

// An example reference implementation of a random password generator.
package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/kloeckner-i/can-haz-password/password"
)

const (
	unambiguousLetters = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz"
	unambiguousDigits  = "23456789"
	// Widely compatible and unambiguous characters.
	specialCharacters = "_-@!*."
)

// A simple command line tool / sanity check for exploring the behavior of the generator library.
func main() {
	passwordLength := flag.Int("length", 8, "minimum length of the generated password")
	includeSpecialCharacters := flag.Bool("special", true, "include special characters")

	flag.Parse()

	generator := password.NewGenerator(&demoPasswordRule{
		length:            passwordLength,
		specialCharacters: includeSpecialCharacters,
	})

	password, err := generator.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Println(password)
}

type demoPasswordRule struct {
	length            *int
	specialCharacters *bool
}

func (r *demoPasswordRule) Config() *password.Configuration {
	passwordLength := *r.length
	classes := []password.CharacterClassConfiguration{
		// Typically the minimums would be constants, in this case due to varying minimum lengths
		// we set them as relative percentages of the total length.
		{Characters: unambiguousLetters, Minimum: int(math.Ceil(float64(passwordLength) * 0.5))},
		{Characters: unambiguousDigits, Minimum: int(math.Ceil(float64(passwordLength) * 0.33))},
	}

	if *r.specialCharacters {
		classes = append(classes, password.CharacterClassConfiguration{
			Characters: specialCharacters,
			Minimum:    int(math.Ceil(float64(passwordLength) * 0.17))})
	}

	return &password.Configuration{
		Length:           passwordLength,
		CharacterClasses: classes,
	}
}

func (r *demoPasswordRule) Valid(_ []rune) bool {
	return true
}
