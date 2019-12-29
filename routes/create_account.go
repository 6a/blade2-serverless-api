package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/6a/blade-ii-api/database"
	"github.com/6a/blade-ii-api/errors"
	"github.com/6a/blade-ii-api/types"
	"github.com/6a/blade-ii-api/utility"
	"github.com/6a/blade-ii-api/validation"
	"github.com/aws/aws-lambda-go/events"
)

// CreateAccount creates an account
func CreateAccount(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	ucr := types.UserCreationRequest{}

	err = json.Unmarshal([]byte(request.Body), &ucr)
	if err != nil {
		r.StatusCode = 400
		r.Message = utility.MakeMessageBody(
			errors.Make(
				errors.RequestMarshalError,
				"Could not unmarshal message body",
			),
		)

		return r, nil
	}

	fieldsValid, message := validateFields(ucr)
	if !fieldsValid {
		r.StatusCode = 400
		r.Message = message

		return r, nil
	}

	handleLengthValid, message := validateHandleLength(*ucr.Handle)
	if !handleLengthValid {
		r.StatusCode = 400
		r.Message = message

		return r, nil
	}

	handleCharactersValid, message := validateHandleFormat(*ucr.Handle)
	if !handleCharactersValid {
		r.StatusCode = 400
		r.Message = message

		return r, nil
	}

	emailCharactersValid, message := validateEmailFormat(*ucr.Email)
	if !emailCharactersValid {
		r.StatusCode = 400
		r.Message = message

		return r, nil
	}

	passwordLengthValid, message := validatePasswordFormat(*ucr.Password)
	if !passwordLengthValid {
		r.StatusCode = 400
		r.Message = message

		return r, nil
	}

	emailConfirmationToken, err := database.CreateUser(*ucr.Handle, *ucr.Email, *ucr.Password)

	log.Printf(emailConfirmationToken)

	return r, nil
}

func validateFields(target types.UserCreationRequest) (ok bool, message string) {
	var field string
	var err uint16
	var expectedType string

	if target.Handle == nil {
		field = "handle"
		err = errors.HandleMissingOrWrongType
		expectedType = "string"
	} else if target.Email == nil {
		field = "email"
		err = errors.EmailMissingOrWrongType
		expectedType = "string"
	} else if target.Password == nil {
		field = "password"
		err = errors.PasswordMissingOrWrongType
		expectedType = "string"
	} else {
		ok = true
	}

	if len(field) != 0 {
		message = utility.MakeMessageBody(
			errors.Make(
				err,
				fmt.Sprintf("Field (%v of type %v) not found, or could not be parsed due to incorrect typing", field, expectedType),
			),
		)
	}

	return ok, message
}

func validateHandleLength(handle string) (valid bool, message string) {
	min, max := validation.UsernameMinLength, validation.UsernameMaxLength
	handleLength := len([]rune(handle))
	valid = handleLength >= min && handleLength <= max

	if !valid {
		message = utility.MakeMessageBody(
			errors.Make(
				errors.HandleLength,
				fmt.Sprintf("handle must be between %v and %v characters", min, max),
			),
		)
	}

	return valid, message
}

func validatePasswordFormat(password string) (valid bool, message string) {
	valid = validation.ValidPasswordChars.MatchString(password)

	if !valid {
		message = utility.MakeMessageBody(
			errors.Make(
				errors.PasswordFormat,
				"Passwords can only contain printable ASCII characters",
			),
		)
	} else {
		passwordLength := len([]rune(password))

		if passwordLength <= validation.PasswordMinLengthLong {
			meetsMinLengthRequirement := passwordLength <= validation.PasswordMinLengthLong
			containsAtLeastOneNumber := validation.NumberAtAnyPosition.MatchString(password)
			containsAtLeastOneLowerCaseChar := validation.LowerCaseAtAnyPosition.MatchString(password)

			if !meetsMinLengthRequirement || !containsAtLeastOneNumber || !containsAtLeastOneLowerCaseChar {
				message = utility.MakeMessageBody(
					errors.Make(
						errors.PasswordComplexityInsufficient,
						"Passwords does not meet minimum complexity requirements",
					),
				)
			}
		}
	}

	return valid, message
}

func validateHandleFormat(handle string) (valid bool, message string) {
	valid = validation.NoSpaceAtStart.MatchString(handle) && validation.ValidUsernameRegex.MatchString(handle)

	if !valid {
		message = utility.MakeMessageBody(
			errors.Make(
				errors.HandleFormat,
				"Handles can only contain full-width japanese characters, half-width alphanumerical characters and certain symbols",
			),
		)
	}

	return valid, message
}

func validateEmailFormat(email string) (valid bool, message string) {
	valid = validation.ValidEmail.MatchString(email)

	if !valid {
		message = utility.MakeMessageBody(
			errors.Make(
				errors.EmailFormat,
				"Email address format is invalid",
			),
		)
	}

	return valid, message
}
