// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package types defines types and contstants for this application.
package types

// Token is a uint16 typedef used for the enumeration of different token types used by this application.
type Token byte

// Token Types.
const (
	AuthToken Token = iota
	EmailConfirmationToken
	PasswordResetToken
	RefreshToken
)

// String is a helper function that returns the token as a string.
func (token Token) String() string {

	// Create an array of all the possible strings.
	types := [...]string{
		"auth",
		"email_confirmation",
		"password_reset",
		"refresh",
	}

	// If the token's value is outside of the accepted range, return a default value.
	if token < AuthToken || token > RefreshToken {
		return "unknown"
	}

	// Get the string from the types array that corresponds with this token.
	return types[token]
}
