package routes

import (
	"fmt"
	"strings"

	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/internal/validation"
)

func packageGenericError(httpCode types.HTTPCode, b2code types.B2ResultCode, err error) (response types.LambdaResponse) {
	return types.MakeLambdaResponse(httpCode, b2code, err.Error())
}

func validateMMRUpdateFields(target types.MMRUpdateRequest) (ok bool, code types.B2ResultCode, payload string) {
	var field string
	var expectedType string

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
		ok = true
	}

	if len(field) != 0 {
		payload = fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType)
	}

	return ok, code, payload
}

func validateUCRFields(target types.UserCreationRequest) (ok bool, code types.B2ResultCode, payload string) {
	var field string
	var expectedType string

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
		ok = true
	}

	if len(field) != 0 {
		payload = fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType)
	}

	return ok, code, payload
}

func validateHandleLength(handle string) (valid bool, code types.B2ResultCode, payload string) {
	min, max := validation.UsernameMinLength, validation.UsernameMaxLength
	handleLength := len([]rune(handle))
	valid = handleLength >= min && handleLength <= max

	if !valid {
		code = types.HandleLength
		payload = fmt.Sprintf("handle must be between %v and %v characters", min, max)
	}

	return valid, code, payload
}

func validatePasswordFormat(password string) (valid bool, code types.B2ResultCode, payload string) {
	valid = validation.ValidPasswordChars.MatchString(password)

	if !valid {
		code = types.PasswordFormat
		payload = "Passwords can only contain printable ASCII characters"
	} else {
		passwordLength := len([]rune(password))

		if passwordLength < validation.PasswordMinLengthLong {
			meetsMinLengthRequirement := passwordLength < validation.PasswordMinLengthLong
			containsAtLeastOneNumber := validation.NumberAtAnyPosition.MatchString(password)
			containsAtLeastOneLowerCaseChar := validation.LowerCaseAtAnyPosition.MatchString(password)

			if !meetsMinLengthRequirement || !containsAtLeastOneNumber || !containsAtLeastOneLowerCaseChar {
				code = types.PasswordComplexityInsufficient
				payload = "Password does not meet minimum complexity requirements"
			}
		}
	}

	return valid, code, payload
}

func validateHandleFormat(handle string) (valid bool, code types.B2ResultCode, payload string) {
	valid = validation.NoSpaceAtStart.MatchString(handle)
	if !valid {
		code = types.HandleSpaceAtStart
		payload = "Handles cannot start with a space"
		return valid, code, payload
	}

	valid = validation.ValidUsernameRegex.MatchString(handle)
	if !valid {
		code = types.HandleFormat
		payload = "Handles can only contain full-width japanese characters, half-width alphanumerical characters and certain symbols"
		return valid, code, payload
	}

	return valid, code, payload
}

func validateEmailFormat(email string) (valid bool, code types.B2ResultCode, payload string) {
	valid = validation.ValidEmail.MatchString(email)

	if !valid {
		code = types.EmailFormat
		payload = "Email address format is invalid"
	}

	return valid, code, payload
}

func packageCreateAccountError(err error) (response types.LambdaResponse) {
	code := types.Success
	payload := ""

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

	return types.MakeLambdaResponse(400, code, payload)
}

func packageMMRUpdateError(DBID uint64, mmrerror error) (response types.LambdaResponse) {
	code := types.DatabaseError
	payload := ""

	if strings.Contains(mmrerror.Error(), "no rows in result set") {
		payload = fmt.Sprintf("MMR update error - player [ %v ] not found", DBID)
	} else {
		payload = fmt.Sprintf("Unknown database error: %v", mmrerror.Error())
	}

	return types.MakeLambdaResponse(400, code, payload)
}

func packageProfileGetError(publicID string, pgError error) (response types.LambdaResponse) {
	code := types.DatabaseError
	payload := pgError.Error()

	return types.MakeLambdaResponse(400, code, payload)
}
