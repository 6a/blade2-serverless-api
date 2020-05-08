// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/email"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/pkg/profanity"
	"github.com/aws/aws-lambda-go/events"
)

// CreateAccount creates an account based on the data contained in the message body.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func CreateAccount(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Attempt to parse the request body as a UserCreationRequest struct.
	ucr := types.UserCreationRequest{}
	err = json.Unmarshal([]byte(request.Body), &ucr)
	if err != nil {
		r = packageGenericError(400, types.RequestMarshalError, err)
		return r, nil
	}

	// Check to ensure that all the expected fields were present in the JSON
	// body, with the correct format, type etc..
	fieldsValid, code, info := validateUCRFields(ucr)
	if !fieldsValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	// The following validation functions ensure that the provided details are appropriate within the
	// context of this app - performing length, format, profanity checks etc..

	handleLengthValid, code, info := validateHandleLength(*ucr.Handle)
	if !handleLengthValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	handleCharactersValid, code, info := validateHandleFormat(*ucr.Handle)
	if !handleCharactersValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	emailCharactersValid, code, info := validateEmailFormat(*ucr.Email)
	if !emailCharactersValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	passwordLengthValid, code, info := validatePasswordFormat(*ucr.Password)
	if !passwordLengthValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	rude := profanity.ContainsProfanity(*ucr.Handle)
	if rude {
		r = packageGenericError(400, types.HandleRude, errors.New("Handle contains profanity"))
		return r, nil
	}

	// Attempt to create the user. A failure Indicates that there was either a database error,
	// or the user already exists etc..
	emailConfirmationToken, err := database.CreateUser(*ucr.Handle, *ucr.Email, *ucr.Password)
	if err != nil {
		r = packageCreateAccountError(err)
		return r, nil
	}

	// Send the email confirmation to the address specified.
	err = email.SendEmailConfirmation(*ucr.Email, *ucr.Handle, emailConfirmationToken)
	if err != nil {
		r = packageGenericError(500, types.EmailSendFailure, err)
		return r, nil
	}

	// Create a message body containing the return data for this API call - in this case
	// the handle for the user that was created.
	createAccountResponse := types.CreateAccountResponsePayload{
		Handle: *ucr.Handle,
	}

	// Package the return payload in a lambda response.
	r = types.MakeLambdaResponse(200, types.Success, createAccountResponse)

	return r, nil
}
