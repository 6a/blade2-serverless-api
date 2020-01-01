package routes

import (
	"context"

	"github.com/6a/blade-ii-api/internal/auth"
	"github.com/6a/blade-ii-api/internal/b2error"
	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/settings"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/pkg/rid"
	"github.com/aws/aws-lambda-go/events"
)

const handleParameter = "user"

// GetAuthToken validates credentials and returns an auth token on success
func GetAuthToken(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	handle, password, err := auth.ExtractCredentials(request.Headers)
	if err != nil {
		r.StatusCode = 401
		r.Body = packageCredentialsExtractionError(err)

		return r, nil
	}

	err = auth.ValidatePathAndHandle(request.PathParameters[handleParameter], handle)
	if err != nil {
		r.StatusCode = 403
		r.Body = packageCredentialsPathAndHandleValidationError(err)

		return r, nil
	}

	err = database.ValidateCredentials(handle, password)
	if err != nil {
		r.StatusCode = 403
		r.Body = packageCredentialsValidationError(err)

		return r, nil
	}

	id, publicID, err := database.GetIDs(handle)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageGetIDError(err)

		return r, nil
	}

	authToken, err := rid.RandomString(settings.AuthTokenLength)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageTokenGenerationError(err)

		return r, nil
	}

	err = database.SetToken(id, types.AuthToken, authToken, settings.AuthTokenLifetime)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageSetTokenError(err)

		return r, nil
	}

	refreshToken, err := rid.RandomString(settings.RefreshTokenLength)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageTokenGenerationError(err)

		return r, nil
	}

	err = database.SetToken(id, types.RefreshToken, refreshToken, settings.RefreshTokenLifetime)
	if err != nil {
		r.StatusCode = 500
		r.Body = packageSetTokenError(err)

		return r, nil
	}

	r.StatusCode = 200
	r.Body = types.GetAuthTokenResponse{
		PublicID:     publicID,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
	}.ToJSON()

	return r, nil
}

func packageCredentialsExtractionError(err error) (body string) {
	body = b2error.Make(
		b2error.AuthHeaderMissing,
		err.Error(),
	).ToJSON()

	return body
}

func packageCredentialsPathAndHandleValidationError(err error) (body string) {
	body = b2error.Make(
		b2error.AuthInsufficientPermissions,
		err.Error(),
	).ToJSON()

	return body
}

func packageCredentialsValidationError(err error) (body string) {
	body = b2error.Make(
		b2error.AuthUsernameOrPasswordIncorrect,
		"Username or password is incorrect", // We dont send the error to reduce the risk of someone working out how the process works
	).ToJSON()

	return body
}

func packageTokenGenerationError(err error) (body string) {
	body = b2error.Make(
		b2error.CryptoRandomError,
		err.Error(),
	).ToJSON()

	return body
}

func packageGetIDError(err error) (body string) {
	body = b2error.Make(
		b2error.DatabaseError,
		err.Error(),
	).ToJSON()

	return body
}

func packageSetTokenError(err error) (body string) {
	body = b2error.Make(
		b2error.DatabaseError,
		err.Error(),
	).ToJSON()

	return body
}
