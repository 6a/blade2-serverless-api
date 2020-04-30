package routes

import (
	"context"

	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

// UpdateMMR updates the mmr for the two specified clients, based on their current MMR, and which client won
func UpdateMMR(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	return r, nil
}

// // CreateAccount creates an account
// func CreateAccount(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
// 	ucr := types.UserCreationRequest{}

// 	err = json.Unmarshal([]byte(request.Body), &ucr)
// 	if err != nil {
// 		r = packageGenericError(400, types.RequestMarshalError, err)
// 		return r, nil
// 	}

// 	fieldsValid, code, payload := validateFields(ucr)
// 	if !fieldsValid {
// 		r = types.MakeLambdaResponse(400, code, payload)
// 		return r, nil
// 	}

// 	handleLengthValid, code, payload := validateHandleLength(*ucr.Handle)
// 	if !handleLengthValid {
// 		r = types.MakeLambdaResponse(400, code, payload)
// 		return r, nil
// 	}

// 	handleCharactersValid, code, payload := validateHandleFormat(*ucr.Handle)
// 	if !handleCharactersValid {
// 		r = types.MakeLambdaResponse(400, code, payload)
// 		return r, nil
// 	}

// 	emailCharactersValid, code, payload := validateEmailFormat(*ucr.Email)
// 	if !emailCharactersValid {
// 		r = types.MakeLambdaResponse(400, code, payload)
// 		return r, nil
// 	}

// 	passwordLengthValid, code, payload := validatePasswordFormat(*ucr.Password)
// 	if !passwordLengthValid {
// 		r = types.MakeLambdaResponse(400, code, payload)
// 		return r, nil
// 	}

// 	rude := profanity.ContainsProfanity(*ucr.Handle)
// 	if rude {
// 		r = packageGenericError(400, types.HandleRude, errors.New("Handle contains profanity"))
// 		return r, nil
// 	}

// 	emailConfirmationToken, err := database.CreateUser(*ucr.Handle, *ucr.Email, *ucr.Password)
// 	if err != nil {
// 		r = packageCreateAccountError(err)
// 		return r, nil
// 	}

// 	err = email.SendEmailConfirmation(*ucr.Email, *ucr.Handle, emailConfirmationToken)
// 	if err != nil {
// 		r = packageGenericError(500, types.EmailSendFailure, err)
// 		return r, nil
// 	}

// 	successPayload := types.CreateAccountResponsePayload{
// 		Handle: *ucr.Handle,
// 	}

// 	r = types.MakeLambdaResponse(200, types.Success, successPayload)

// 	return r, nil
// }
