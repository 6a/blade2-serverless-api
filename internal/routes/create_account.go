package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/6a/blade-ii-api/internal/b2error"
	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/email"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/internal/validation"
	"github.com/aws/aws-lambda-go/events"
)

// CreateAccount creates an account
func CreateAccount(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	ucr := types.UserCreationRequest{}

	err = json.Unmarshal([]byte(request.Body), &ucr)
	if err != nil {
		r.StatusCode = 400
		r.Body = b2error.Make(
			b2error.RequestMarshalError,
			"Could not unmarshal message body",
		).ToJSON()

		return r, nil
	}

	fieldsValid, body := validateFields(ucr)
	if !fieldsValid {
		r.StatusCode = 400
		r.Body = body

		return r, nil
	}

	handleLengthValid, body := validateHandleLength(*ucr.Handle)
	if !handleLengthValid {
		r.StatusCode = 400
		r.Body = body

		return r, nil
	}

	handleCharactersValid, body := validateHandleFormat(*ucr.Handle)
	if !handleCharactersValid {
		r.StatusCode = 400
		r.Body = body

		return r, nil
	}

	emailCharactersValid, body := validateEmailFormat(*ucr.Email)
	if !emailCharactersValid {
		r.StatusCode = 400
		r.Body = body

		return r, nil
	}

	passwordLengthValid, body := validatePasswordFormat(*ucr.Password)
	if !passwordLengthValid {
		r.StatusCode = 400
		r.Body = body

		return r, nil
	}

	emailConfirmationToken, err := database.CreateUser(*ucr.Handle, *ucr.Email, *ucr.Password)
	if err != nil {
		r.StatusCode = 400
		r.Body = packageCreateAccountError(err)

		return r, nil
	}

	err = email.SendEmailConfirmation(*ucr.Email, *ucr.Handle, emailConfirmationToken)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageEmailError(err)

		return r, nil
	}

	r.StatusCode = 200
	r.Body = b2error.Make(b2error.None, *ucr.Handle).ToJSON()

	return r, nil
}

func validateFields(target types.UserCreationRequest) (ok bool, body string) {
	var field string
	var err uint16
	var expectedType string

	if target.Handle == nil {
		field = "handle"
		err = b2error.HandleMissingOrWrongType
		expectedType = "string"
	} else if target.Email == nil {
		field = "email"
		err = b2error.EmailMissingOrWrongType
		expectedType = "string"
	} else if target.Password == nil {
		field = "password"
		err = b2error.PasswordMissingOrWrongType
		expectedType = "string"
	} else {
		ok = true
	}

	if len(field) != 0 {
		body = b2error.Make(
			err,
			fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType),
		).ToJSON()
	}

	return ok, body
}

func validateHandleLength(handle string) (valid bool, body string) {
	min, max := validation.UsernameMinLength, validation.UsernameMaxLength
	handleLength := len([]rune(handle))
	valid = handleLength >= min && handleLength <= max

	if !valid {
		body = b2error.Make(
			b2error.HandleLength,
			fmt.Sprintf("handle must be between %v and %v characters", min, max),
		).ToJSON()
	}

	return valid, body
}

func validatePasswordFormat(password string) (valid bool, body string) {
	valid = validation.ValidPasswordChars.MatchString(password)

	if !valid {
		body = b2error.Make(
			b2error.PasswordFormat,
			"Passwords can only contain printable ASCII characters",
		).ToJSON()
	} else {
		passwordLength := len([]rune(password))

		if passwordLength < validation.PasswordMinLengthLong {
			meetsMinLengthRequirement := passwordLength < validation.PasswordMinLengthLong
			containsAtLeastOneNumber := validation.NumberAtAnyPosition.MatchString(password)
			containsAtLeastOneLowerCaseChar := validation.LowerCaseAtAnyPosition.MatchString(password)

			if !meetsMinLengthRequirement || !containsAtLeastOneNumber || !containsAtLeastOneLowerCaseChar {
				valid = false
				body = b2error.Make(
					b2error.PasswordComplexityInsufficient,
					"Password does not meet minimum complexity requirements",
				).ToJSON()
			}
		}
	}

	return valid, body
}

func validateHandleFormat(handle string) (valid bool, body string) {
	valid = validation.NoSpaceAtStart.MatchString(handle)
	if !valid {
		body = b2error.Make(
			b2error.HandleSpaceAtStart,
			"Handles cannot start with a space",
		).ToJSON()
		return valid, body
	}

	valid = validation.ValidUsernameRegex.MatchString(handle)
	if !valid {
		body = b2error.Make(
			b2error.HandleFormat,
			"Handles can only contain full-width japanese characters, half-width alphanumerical characters and certain symbols",
		).ToJSON()
	}

	return valid, body
}

func validateEmailFormat(email string) (valid bool, body string) {
	valid = validation.ValidEmail.MatchString(email)

	if !valid {
		body = b2error.Make(
			b2error.EmailFormat,
			"Email address format is invalid",
		).ToJSON()
	}

	return valid, body
}

func packageCreateAccountError(err error) (body string) {
	if strings.Contains(err.Error(), "Error 1062") {
		if strings.Contains(err.Error(), "email_UNIQUE") {
			body = b2error.Make(
				b2error.EmailAlreadyInUse,
				"Email address already in use",
			).ToJSON()
		} else if strings.Contains(err.Error(), "handle_UNIQUE") {
			body = b2error.Make(
				b2error.HandleAlreadyInUse,
				"Handle already in use",
			).ToJSON()
		}
	} else {
		body = b2error.Make(
			b2error.DatabaseError,
			fmt.Sprintf("Unknown database error: %v", err.Error()),
		).ToJSON()
	}

	return body
}

func packageEmailError(err error) (body string) {
	return body
}
