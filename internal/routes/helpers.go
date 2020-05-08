// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"fmt"
	"strings"

	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/internal/validation"
)

// packageGenericError creates a lambda that will result in a HTTP response with the specified HTTP status code. The
// message body will contain a JSON encoded body containing the specified code, and the error as a string.
func packageGenericError(httpCode types.HTTPCode, b2code types.B2ResultCode, err error) (response types.LambdaResponse) {
	return types.MakeLambdaResponse(httpCode, b2code, err.Error())
}

// validateMMRUpdateFields returns true if the fields in an MMR update request are valid. If null, this would
// suggest that the JSON string parsing process failed, due to a field being missage or of an incorrect type.
// Returns true when the request is considered to be valid, and returns a result code and some relevant info
// if invalid.
func validateMMRUpdateFields(target types.MMRUpdateRequest) (ok bool, code types.B2ResultCode, info string) {

	// Declare some variables to store the field name and type, for building the info string
	// when an error is detected.
	var field string
	var expectedType string

	// Check each struct member to see if they are nil - which would indicate that there was an error, and
	// the request is invalid. Set valus for the error code, as well as field and expected type.
	if target.Player1ID == nil {
		field = "player1id"
		code = types.Player1IDMissingOrWrongType
		expectedType = "uint64"
	} else if target.Player2ID == nil {
		field = "player2id"
		code = types.Player2IDMissingOrWrongType
		expectedType = "uint64"
	} else if target.Winner == nil {
		field = "winner"
		code = types.WinnerMissingOrWrongType
		expectedType = "uint8"
	} else {

		// If there was no error, set the return boolean to true, so the caller is aware that the specified update
		// request was valid.
		ok = true
	}

	// If the field variable has a value, then there was at least one error - so create the info string to be returned.
	if len(field) != 0 {
		info = fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType)
	}

	return ok, code, info
}

// validateUCRFields returns true if the fields in an user creation request are valid. If null, this would
// suggest that the JSON string parsing process failed, due to a field being missage or of an incorrect type.
// Returns true when the request is considered to be valid, and returns a result code and some relevant info
// if invalid.
func validateUCRFields(target types.UserCreationRequest) (ok bool, code types.B2ResultCode, info string) {

	// Declare some variables to store the field name and type, for building the info string
	// when an error is detected.
	var field string
	var expectedType string

	// Check each struct member to see if they are nil - which would indicate that there was an error, and
	// the request is invalid. Set valus for the error code, as well as field and expected type.
	if target.Handle == nil {
		field = "handle"
		code = types.HandleMissingOrWrongType
		expectedType = "string"
	} else if target.Email == nil {
		field = "email"
		code = types.EmailMissingOrWrongType
		expectedType = "string"
	} else if target.Password == nil {
		field = "password"
		code = types.PasswordMissingOrWrongType
		expectedType = "string"
	} else {

		// If there was no error, set the return boolean to true, so the caller is aware that the specified update
		// request was valid.
		ok = true
	}

	// If the field variable has a value, then there was at least one error - so create the info string to be returned.
	if len(field) != 0 {
		info = fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType)
	}

	return ok, code, info
}

// validateMMRUpdateFields returns true if a handle meets the requirements for this application.
func validateHandleLength(handle string) (valid bool, code types.B2ResultCode, info string) {

	// For convenience, store the minimum and maximum values in two local variables.
	min, max := validation.UsernameMinLength, validation.UsernameMaxLength

	// Determine the length of the handle. Note that the handle is first converted in a rune array,
	// to ensure that characters wider than 1 byte (such as unicode characters) are still only
	// considered as a single character.
	handleLength := len([]rune(handle))

	// Check to see if the handle length is within the size range expected by the application.
	valid = handleLength >= min && handleLength <= max

	// If invalid, set the return code and the info string for this function.
	if !valid {
		code = types.HandleLength
		info = fmt.Sprintf("handle must be between %v and %v characters", min, max)
	}

	return valid, code, info
}

// validatePasswordFormat returns true if the password meets the requirements for
// this application. Inspired by GitHub's password requirements:
// https://help.github.com/en/github/authenticating-to-github/creating-a-strong-password
func validatePasswordFormat(password string) (valid bool, code types.B2ResultCode, info string) {

	// Ensure that the password only contains valid characters - a return value of false indicates
	// that it did not.
	valid = validation.ValidPasswordChars.MatchString(password)

	// Exit early if the password contained invalid characters.
	if !valid {
		code = types.PasswordFormat
		info = "Passwords can only contain printable ASCII characters"
	} else {

		// Determine the length of the password. Note that the handle is first converted in a rune array,
		// to ensure that characters wider than 1 byte (such as unicode characters) are still only
		// considered as a single character.
		passwordLength := len([]rune(password))

		// If the password does not meet the minimum length required to be considered valid REGARDLESS of
		// the characters used... If it IS long enough, then it's considered valid.
		if passwordLength < validation.PasswordMinLengthLong {

			// Check the following requirements:

			// Check that the password meets the minimum length requirement.
			meetsMinLengthRequirement := passwordLength < validation.PasswordMinLengthLong

			// Check that that the password contains at least one number.
			containsAtLeastOneNumber := validation.NumberAtAnyPosition.MatchString(password)

			// Check that the password contains at leaast one lower case character.
			containsAtLeastOneLowerCaseChar := validation.LowerCaseAtAnyPosition.MatchString(password)

			// If any of the following checks returned true, the password was invalid in some way - set the
			// return code and info to appropriate values.
			if !meetsMinLengthRequirement || !containsAtLeastOneNumber || !containsAtLeastOneLowerCaseChar {
				code = types.PasswordComplexityInsufficient
				info = "Password does not meet minimum complexity requirements"
			}
		}
	}

	return valid, code, info
}

// validateHandleFormat returns true if the specified handle meets the requirements for this application.
func validateHandleFormat(handle string) (valid bool, code types.B2ResultCode, info string) {

	// Check to see if there is a space at the start.
	valid = validation.NoSpaceAtStart.MatchString(handle)
	if !valid {
		code = types.HandleSpaceAtStart
		info = "Handles cannot start with a space"
		return valid, code, info
	}

	// Check to see if there are any invalid characters.
	valid = validation.ValidUsernameRegex.MatchString(handle)
	if !valid {
		code = types.HandleFormat
		info = "Handles can only contain full-width japanese characters, half-width alphanumerical characters and certain symbols"
		return valid, code, info
	}

	return valid, code, info
}

// validateEmailFormat returns true if the specified email address meets the requirements for this application.
func validateEmailFormat(email string) (valid bool, code types.B2ResultCode, info string) {

	// Check to see if the email matches the email validation regex pattern.
	valid = validation.ValidEmail.MatchString(email)
	if !valid {
		code = types.EmailFormat
		info = "Email address format is invalid"
	}

	return valid, code, info
}

// packageCreateAccountError creates a lamda response based on the specified account creation error.
func packageCreateAccountError(err error) (response types.LambdaResponse) {

	// Declare variables for the code and payload, to be set depending on the error.
	code := types.Success
	payload := ""

	// Depending on the contents of the error, determine the code and message body.
	// Unexpected errors are packaged with a generic message.
	if strings.Contains(err.Error(), "Error 1062") {
		if strings.Contains(err.Error(), "email_UNIQUE") {
			code = types.EmailAlreadyInUse
			payload = "Email address already in use"
		} else if strings.Contains(err.Error(), "handle_UNIQUE") {
			code = types.HandleAlreadyInUse
			payload = "Handle already in use"
		}
	} else {
		code = types.DatabaseError
		payload = fmt.Sprintf("Unknown database error: %v", err.Error())
	}

	// Package and return the code and message payload as a lambda response.
	return types.MakeLambdaResponse(400, code, payload)
}

// packageMMRUpdateError creates a lamda response based on the specified MMR update error.
func packageMMRUpdateError(DBID uint64, mmrerror error) (response types.LambdaResponse) {

	// Declare variables for the code and payload, to be set depending on the error.
	code := types.DatabaseError
	payload := ""

	// Depending on the contents of the error, determine the code and message body.
	// Unexpected errors are packaged with a generic message.
	if strings.Contains(mmrerror.Error(), "no rows in result set") {
		payload = fmt.Sprintf("MMR update error - player [ %v ] not found", DBID)
	} else {
		payload = fmt.Sprintf("Unknown database error: %v", mmrerror.Error())
	}

	// Package and return the code and message payload as a lambda response.
	return types.MakeLambdaResponse(400, code, payload)
}

func packageProfileGetError(publicID string, pgError error) (response types.LambdaResponse) {

	// Declare variables for the code and payload, to be set depending on the error.
	code := types.DatabaseError
	payload := fmt.Sprintf("Unknown database error: %v", pgError.Error())

	// No filtering implemented - all errors are returned as a database error with
	// the raw error in the payload.
	// TODO implement proper error messages.

	// Package and return the code and message payload as a lambda response.
	return types.MakeLambdaResponse(400, code, payload)
}

func packageLeaderboardsError(lbError error) (response types.LambdaResponse) {

	// Declare variables for the code and payload, to be set depending on the error.
	code := types.DatabaseError
	payload := fmt.Sprintf("Unknown database error: %v", lbError.Error())

	// No filtering implemented - all errors are returned as a database error with
	// the raw error in the payload.
	// TODO implement proper error messages.

	// Package and return the code and message payload as a lambda response.
	return types.MakeLambdaResponse(400, code, payload)
}

func packageMatchHistoryError(mhError error) (response types.LambdaResponse) {

	// Declare variables for the code and payload, to be set depending on the error.
	code := types.DatabaseError
	payload := fmt.Sprintf("Unknown database error: %v", mhError.Error())

	// No filtering implemented - all errors are returned as a database error with
	// the raw error in the payload.
	// TODO implement proper error messages.

	// Package and return the code and message payload as a lambda response.
	return types.MakeLambdaResponse(400, code, payload)
}
