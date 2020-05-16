// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"context"
	"errors"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

const publicIDParameterKey = "pid"

// GetProfile returns the profile data for the user specified by the public ID in the path /profiles/{publicID}.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func GetProfile(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Check for the existence of, and then get the value for the "pid" path parameter.
	var pid string
	if _, ok := request.PathParameters[publicIDParameterKey]; ok {
		pid = request.PathParameters[publicIDParameterKey]
	} else {
		r = packageGenericError(400, types.ProfileGetPublicIDMising, errors.New("Public ID parameter missing"))
		return r, nil
	}

	// Attempt to get the database ID for the user specified by public ID.
	DBID, err := database.GetDatabaseID(pid)
	if err != nil {
		r = packageGenericError(404, types.ProfileGetPublicIDNotFound, errors.New("No matching profile found"))
		return r, nil
	}

	// Attempt to get the profile data for the user. This will be passed directly into the lambda response
	// make function, to be packaged as a JSON string in the message body.
	profile, err := database.GetProfile(DBID)
	if err != nil {
		r = packageProfileGetError(pid, err)
		return r, nil
	}

	// Package the return payload in a lambda response.
	r = types.MakeLambdaResponse(200, types.Success, profile)

	return r, nil
}
