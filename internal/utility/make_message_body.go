package utility

import (
	"encoding/json"

	"github.com/6a/blade-ii-api/internal/errors"
)

// MakeMessageBody takes a B2Error and returns it as a json formatted string
func MakeMessageBody(err errors.B2Error) string {
	// We can blindly marshal without checking for error, as the input should be valid
	jsonMessage, _ := json.Marshal(err)

	return string(jsonMessage)
}
