package b2error

import "encoding/json"

// Code is a typedef for uint16, used for the enumerations of error codes
type Code = uint16

// Error is a container for error data
type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// ToJSON returns a json format string of owning B2Error
func (e Error) ToJSON() (jsonMessageBody string) {
	jsonMessage, _ := json.Marshal(e)
	return string(jsonMessage)
}

// None is used when there is no error
const None Code = iota

// Generic errors
const (
	RequestMarshalError Code = iota + 100
	DatabaseError
)

// Create account handle errors
const (
	HandleMissingOrWrongType Code = iota + 200
	HandleLength
	HandleSpaceAtStart
	HandleFormat
	HandleAlreadyInUse
	HandleRude
)

// Create account email errors
const (
	EmailMissingOrWrongType Code = iota + 300
	EmailFormat
	EmailAlreadyInUse
)

// Create account password errors
const (
	PasswordMissingOrWrongType Code = iota + 400
	PasswordComplexityInsufficient
	PasswordFormat
)

// Make returns a new B2Error object based on the provided argument
func Make(code Code, message string) Error {
	return Error{Code: code, Message: message}
}
