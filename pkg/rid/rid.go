// Adapted from https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/

package rid

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomBytes returns an array of random bytes of n length, generated securely using crypto/rand
func RandomBytes(n int) ([]byte, error) {
	buffer := make([]byte, n)

	// Put values into the buffer, and return any errors
	_, err := rand.Read(buffer)
	if err != nil {
		return nil, err
	}

	// Return the successfully filled buffer
	return buffer, nil
}

// RandomString returns a base64 string of random characters of n length
func RandomString(n int) (string, error) {
	b, err := RandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
