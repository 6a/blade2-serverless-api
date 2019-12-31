package auth

import (
	"encoding/base64"
	"errors"
	"strings"
)

// ExtractCredentials attempts to extract and decode the Authorization header in a set of request headers
func ExtractCredentials(headers map[string]string) (handle string, password string, err error) {
	if authHeader, ok := headers["Authorization"]; ok {
		decodedHeader, err := base64.StdEncoding.DecodeString(strings.Replace(authHeader, "Basic ", "", 1))

		if err != nil {
			return "", "", errors.New("Authorization header could not be decoded as a base64 (standard) string")
		}

		var credentials = strings.Split(string(decodedHeader), ":")
		if len(credentials) != 2 {
			return "", "", errors.New("Authorization header format invalid")
		}

		handle, password := credentials[0], credentials[1]

		return handle, password, nil
	}

	return "", "", errors.New("Authorization header not found")
}

// ValidatePathAndHandle checks if the path parameter matches the expected handle
func ValidatePathAndHandle(authHandle string, pathHandle string) (err error) {
	matches := authHandle == pathHandle

	if !matches {
		err = errors.New("The handle specified in the Authorization header does not match the handle specified in the path")
	}

	return err
}
