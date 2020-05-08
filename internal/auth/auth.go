// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package auth implements provides a helper function for extracting the username and password
// from a Basic Authorization header as specified by (RFC7617).
package auth

import (
	"encoding/base64"
	"errors"
	"strings"
)

const (

	// basicAuthHeaderDelimiter as specified by (RFC7617).
	basicAuthHeaderDelimiter = ":"

	// credentialsArrayExpectedSize is the epected size of the Basic Auth header after being decoded and split.
	credentialsArrayExpectedSize = 2
)

// ExtractCredentials attempts to extract and decode the Basic Authorization header in a set of request headers,
// as specified by (RFC7617).
func ExtractCredentials(headers map[string]string) (handle string, password string, err error) {

	// Check for the existence of the authorization header.
	if authHeader, ok := headers["Authorization"]; ok {

		// Attempt to decode the contents of the basic auth header. This involves removing the suffix "Basic "
		// and then decoding the remainining string using standard base64 decoding.
		decodedHeader, err := base64.StdEncoding.DecodeString(strings.Replace(authHeader, "Basic ", "", 1))

		// An error indicates that the format of the authorization header was invalid, and so an error is returned.
		if err != nil {
			return "", "", errors.New("Authorization header could not be decoded as a base64 (standard) string")
		}

		// The decoded string is split into an array. If the length of the resultant array is not
		// (credentialsArrayExpectedSize), an error is returned.
		var credentials = strings.Split(string(decodedHeader), basicAuthHeaderDelimiter)
		if len(credentials) != credentialsArrayExpectedSize {
			return "", "", errors.New("Authorization header format invalid")
		}

		// Return the parsed credentials.
		return credentials[0], credentials[1], nil
	}

	// If the authorization header was not found, exit with an error.
	return "", "", errors.New("Authorization header not found")
}
