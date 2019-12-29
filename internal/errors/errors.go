package errors

import "encoding/json"

// B2ErrorCode is a typedef for uint16, used for the enumerations of error codes
type B2ErrorCode = uint16

// B2Error is a container for error data
type B2Error struct {
	Code    B2ErrorCode `json:"code"`
	Message string      `json:"message"`
}

// ToJSON returns a json format string of owning B2Error
func (e B2Error) ToJSON() (jsonMessageBody string) {
	jsonMessage, _ := json.Marshal(e)
	return string(jsonMessage)
}

// None is used when there is no error
const None B2ErrorCode = iota

// Generic errors
const (
	RequestMarshalError B2ErrorCode = iota + 100
	DatabaseError
)

// Create account errors
const (
	HandleMissingOrWrongType B2ErrorCode = iota + 200
	HandleLength
	HandleFormat
	HandleAlreadyInUse

	EmailMissingOrWrongType
	EmailFormat
	EmailAlreadyInUse

	PasswordMissingOrWrongType
	PasswordComplexityInsufficient
	PasswordFormat
)

// Make returns a new B2Error object based on the provided argument
func Make(code B2ErrorCode, message string) B2Error {
	return B2Error{Code: code, Message: message}
}
