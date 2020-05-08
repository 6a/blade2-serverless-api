// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package settings is a utility package that contains various app-wide constants.
package settings

const (

	// EmailConfirmationTokenLifetime is the number of hours for which an email confirmation token will be valid.
	EmailConfirmationTokenLifetime = 48

	// EmailConfirmationTokenLength is the length of a generated email confirmation token.
	EmailConfirmationTokenLength = 32

	// AuthTokenLifetime is the number of hours for which an auth token will be valid.
	AuthTokenLifetime = 1

	// AuthTokenLength is the length of a generated auth token.
	AuthTokenLength = 32

	// RefreshTokenLifetime is the number of hours for which an auth refresh token will be valid.
	RefreshTokenLifetime = 12

	// RefreshTokenLength is the length of a generated auth refresh token.
	RefreshTokenLength = 32
)
