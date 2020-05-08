// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"context"
	"errors"

	"github.com/6a/blade-ii-api/internal/auth"
	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/settings"
	"github.com/6a/blade-ii-api/internal/types"

	"github.com/6a/blade-ii-game-server/pkg/rid"
	"github.com/aws/aws-lambda-go/events"
)

// GetAuthToken validates credentials and returns an auth token for the user specified.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func GetAuthToken(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Extract the username and password from the Authorization header.
	handle, password, err := auth.ExtractCredentials(request.Headers)
	if err != nil {
		r = packageGenericError(401, types.AuthHeaderMissing, err)
		return r, nil
	}

	// Check to see if the parsed username and password are valid.
	err = database.ValidateCredentials(handle, password)
	if err != nil {
		r = packageGenericError(403, types.AuthUsernameOrPasswordIncorrect, errors.New("Username or password is incorrect"))
		return r, nil
	}

	// Get the database and public ID for this user.
	id, publicID, err := database.GetIDs(handle)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	// Generate a new auth token.
	authToken, err := rid.RandomString(settings.AuthTokenLength)
	if err != nil {
		r = packageGenericError(500, types.CryptoRandomError, err)
		return r, nil
	}

	// Update the tokens table with the new token for this user.
	err = database.SetToken(id, types.AuthToken, authToken, settings.AuthTokenLifetime)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	// Generate a new refresh token.
	refreshToken, err := rid.RandomString(settings.RefreshTokenLength)
	if err != nil {
		r = packageGenericError(500, types.CryptoRandomError, err)
		return r, nil
	}

	// Update the tokens table with the new token for this user.
	err = database.SetToken(id, types.RefreshToken, refreshToken, settings.RefreshTokenLifetime)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	// Create a message body containing the return data for this API call - in this case the
	// public ID for this user, as well as the generated auth and refresh tokens.
	authResponse := types.AuthResponsePayload{
		PublicID:     publicID,
		AuthToken:    authToken,
		RefreshToken: refreshToken,
	}

	// Package the return payload in a lambda response.
	r = types.MakeLambdaResponse(200, types.Success, authResponse)

	return r, nil
}
