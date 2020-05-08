// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/6a/blade-ii-api/internal/auth"
	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/6a/blade-ii-api/pkg/elo"
	"github.com/aws/aws-lambda-go/events"
)

// UpdateMMR updates the mmr for the two specified clients, based on their current MMR, and which client won.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func UpdateMMR(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Extract the username and password from the Authorization header.
	handle, password, err := auth.ExtractCredentials(request.Headers)
	if err != nil {
		r = packageGenericError(401, types.AuthHeaderMissing, err)
		return r, nil
	}

	// Check to see if the account specified user has the required privilege level to perform this action.
	// Note that this is done before the credentials check, as the credentials check is fairly slow and being able to
	// exit early should reduce server load.
	hasRequiredPrivilege, err := database.HasRequiredPrivilege(handle, database.GameAdminPrivilege)
	if err != nil || !hasRequiredPrivilege {
		r = packageGenericError(403, types.AuthUsernameOrPasswordIncorrect, errors.New("Username or password is incorrect"))
		return r, nil
	}

	// Check to see if the parsed username and password are valid.
	err = database.ValidateCredentials(handle, password)
	if err != nil {
		r = packageGenericError(403, types.AuthUsernameOrPasswordIncorrect, errors.New("Username or password is incorrect"))
		return r, nil
	}

	// Attempt to parse the message body into an MMR update struct.
	mmrur := types.MMRUpdateRequest{}
	err = json.Unmarshal([]byte(request.Body), &mmrur)
	if err != nil {
		r = packageGenericError(400, types.RequestMarshalError, err)
		return r, nil
	}

	// Check to see if the request body format was valid.
	fieldsValid, code, payload := validateMMRUpdateFields(mmrur)
	if !fieldsValid {
		r = types.MakeLambdaResponse(400, code, payload)
		return r, nil
	}

	// Get the match stats for the first player specified in the update request.
	player1MatchStats, err := database.GetMatchStats(*mmrur.Player1ID)
	if err != nil {
		r = packageMMRUpdateError(*mmrur.Player1ID, err)
		return r, nil
	}

	// Get the match stats for the second player specified in the update request.
	player2MatchStats, err := database.GetMatchStats(*mmrur.Player2ID)
	if err != nil {
		r = packageMMRUpdateError(*mmrur.Player2ID, err)
		return r, nil
	}

	// Calculate the new MMR for both players.
	player1MatchStats.MMR, player2MatchStats.MMR = elo.CalculateNewElo(player1MatchStats.MMR, player2MatchStats.MMR, *mmrur.Winner)

	// Update the match stats for both players.
	err = database.UpdateMatchStats(*mmrur.Player1ID, player1MatchStats, *mmrur.Player2ID, player2MatchStats, *mmrur.Winner)
	if err != nil {
		r = packageGenericError(500, types.DatabaseError, err)
		return r, nil
	}

	// Package an empty string in a lambda response - note the status code of 204, a success with no message body.
	r = types.MakeLambdaResponse(204, types.Success, "")

	return r, nil
}
