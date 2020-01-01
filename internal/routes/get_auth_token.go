package routes

import (
	"context"
	"errors"

	"github.com/6a/blade-ii-api/internal/auth"
	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/settings"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/pkg/rid"
	"github.com/aws/aws-lambda-go/events"
)

const handleParameter = "user"

// GetAuthToken validates credentials and returns an auth token on success
func GetAuthToken(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	handle, password, err := auth.ExtractCredentials(request.Headers)
	if err != nil {
		r = packageGenericError(401, types.AuthHeaderMissing, err)
		return r, nil
	}

	err = database.ValidateCredentials(handle, password)
	if err != nil {
		r = packageGenericError(403, types.AuthUsernameOrPasswordIncorrect, errors.New("Username or password is incorrect"))
		return r, nil
	}

	id, publicID, err := database.GetIDs(handle)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	authToken, err := rid.RandomString(settings.AuthTokenLength)
	if err != nil {
		r = packageGenericError(500, types.CryptoRandomError, err)
		return r, nil
	}

	err = database.SetToken(id, types.AuthToken, authToken, settings.AuthTokenLifetime)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	refreshToken, err := rid.RandomString(settings.RefreshTokenLength)
	if err != nil {
		r = packageGenericError(500, types.CryptoRandomError, err)
		return r, nil
	}

	err = database.SetToken(id, types.RefreshToken, refreshToken, settings.RefreshTokenLifetime)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	payload := types.AuthResponsePayload{
		PublicID:     publicID,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
	}

	r = types.MakeLambdaResponse(200, types.Success, payload)

	return r, nil
}
