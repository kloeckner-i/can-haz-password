# A Tour of `Can Haz Password`'s Architecture

## Random Source

`can-haz-password` uses Go's `crypto/rand` as the source of entropy for 
password generation. However due to `crypto/rand`  exposing a limited set of 
methods which excludes floating point operations, we wrap `crypto/rand` to make
it compatible as a `math/rand` random source. The logic for this wrapping
exists in the `rand/source.go` source file. This is a common pattern.

Go intentionally exposes a small set of methods for `crypto/rand` to discourage
misuse (eg. home rolled cryptography). Our wrapping breaks this design feature
and comes with a set of caveats. There's a fairly decent test for the floating
point correctness in the `rand/source_test.go` source file but there are no 
guarantees provided.

## Weighted Random

As we want to be able to generate passwords with varying properties (eg. the
frequency of letters / digits). We need a way to randomly select characters but
with a fixed non-uniform distribution (eg. for a password with many digits the
selection process should more frequently return digits). The implementation for
this weighted random can be found in the `rand/weighted.go` source file.

The algorithm and data structure used for weighted random value generation is
based around an interval tree of accumulated weights. We take an unordered 
collection of entries, and for each entry we insert it into the interval tree
in the interval `total <= x < total + entry_weight` where total is the sum of 
all of the previous weights already stored in the tree. The insertion order is
not important. The end result of this is an interval tree containing a sequence
of immediately adjacent intervals of varying width / weight. 

By generating a random value in the range `0 <= x < total` we guarantee an 
interval exists in the tree which contains the value. The probability that the
random value will fall in any individual interval is equal to the ratio of its
width to the total width of the tree. This is the mechanism that allows us to 
weight/bias the resulting random distribution.

## Password Generation

Passwords are constructed by reading the weighted random stream until one of
two things happen, either we meet the complexity requirements of the password
rule or we exceed the maximum length, which results in us repeating the process.

As each new character is appended to the password, we evaluate whether the 
resulting password would be considered invalid according to its associated rule.
If invalid we will avoid appending to the password, and continue onto the next
value.

Passwords are returned in the order that they were read from the weighted 
random source.

## Reference Implementation

There's an example implementation of a password rule, and generator available
in `cmd/main.go`. This generator uses a custom character set that is designed
to exclude visually ambiguous characters. This is based off practices 
implemented in the BASE58 encoding scheme, and various existing generators such
as 1Password.

The password generator can be run with `make run` and includes several command
line flags that can be used to tune the output, such as minimum password length
and whether to include special characters.

This example should be used as the canonical documentation/implementation
reference, and should remain as clear as possible.

## Advantages

* Passwords preserve ordering.
* Password length and composition is randomized.
* Relatively fast and efficient.
* Conceptually simple (from an architecture standpoint).