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
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

// UpdateAvatar updates the avatar for the client specified by the public ID in the path /profiles/{publicID}/avatar,
// with the avatar specified in the message body { avatar: {Number} }.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func UpdateAvatar(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Check for the existence of, and then get the value for the "pid" path parameter.
	var pid string
	if _, ok := request.PathParameters[publicIDParameterKey]; ok {
		pid = request.PathParameters[publicIDParameterKey]
	} else {
		r = packageGenericError(400, types.ProfileGetPublicIDMising, errors.New("Public ID parameter missing"))
		return r, nil
	}

	// Attempt to parse the request body as an AvatarUpdateRequest struct.
	aur := types.AvatarUpdateRequest{}
	err = json.Unmarshal([]byte(request.Body), &aur)
	if err != nil {
		r = packageGenericError(400, types.RequestMarshalError, err)
		return r, nil
	}

	// Check to ensure that all the expected fields were present in the JSON
	// body, with the correct format, type etc..
	fieldsValid, code, info := validateAURFields(aur)
	if !fieldsValid {
		r = types.MakeLambdaResponse(400, code, info)
		return r, nil
	}

	// Check that the avatar ID is within the valid range (don't check < 0 as the argument is a uint)
	if *aur.Avatar > 9 {
		r = packageGenericError(400, types.ProfileAvatarUpdateAvatarValueInvalid, errors.New("Avatar must be an int between 0 and 9 inclusive"))
		return r, nil
	}

	// Check that the auth token is valid for this user.
	err = database.CheckAuthToken(pid, *aur.AuthToken)
	if err != nil {
		r = packageAuthTokenCheckError(pid, err)
		return r, nil
	}

	err = database.UpdateAvatar(pid, *aur.Avatar)
	if err != nil {
		r = packageGenericError(404, types.DatabaseError, errors.New("Failed to update avatar"))
		return r, nil
	}

	// Package the return payload in a lambda response.
	r = types.MakeLambdaResponse(204, types.Success, "")

	return r, nil
}
