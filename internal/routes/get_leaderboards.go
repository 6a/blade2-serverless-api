// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes implements various endpoints for the Blade II REST API.
package routes

import (
	"context"
	"errors"
	"strconv"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

const queryParamFrom string = "from"
const queryParamCount string = "count"
const queryParamPublicID string = "pid"
const maxResultsSize uint64 = 100

// GetLeaderboards returns the leaderboard data for the range specified in the query param (from, count), with an
// extra member containing the row for the user specified by the (pid) public ID query param. If pid is unspecified,
// the user member is returned with default values.
//
// Errors will never be returned, and instead will be handled by returning a response with a suitable HTTP status
// code (RFC 7231).
func GetLeaderboards(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {

	// Attempt to fetch and parse each query param:
	// - from: uint64
	// - count: uint64
	// - pid: string [ optional ]

	// Check for the existence of, and then get the value for the "from" query parameter.
	var from string
	if _, ok := request.QueryStringParameters[queryParamFrom]; ok {
		from = request.QueryStringParameters[queryParamFrom]
	} else {
		r = packageGenericError(400, types.LeaderboardsRangeFromMissing, errors.New("'from' query param missing"))
		return r, nil
	}

	// Attempt to parse the "from" string as a uint64.
	fromInt, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		r = packageGenericError(400, types.LeaderboardsRangeFromInvalid, errors.New("'from' query param invalid"))
		return r, nil
	}

	// Check for the existence of, and then get the value for the "count" query parameter.
	var count string
	if _, ok := request.QueryStringParameters[queryParamCount]; ok {
		count = request.QueryStringParameters[queryParamCount]
	} else {
		r = packageGenericError(400, types.LeaderboardsRangeCountMissing, errors.New("'count' query param missing"))
		return r, nil
	}

	// Attempt to parse the "count" string as a uint64.
	countInt, err := strconv.ParseUint(count, 10, 64)
	if err != nil {
		r = packageGenericError(400, types.LeaderboardsRangeCountInvalid, errors.New("'count' query param invalid"))
		return r, nil
	}

	// Check to ensure that the number of results requested does not exceed the maximum - to avoid someone requesting
	// the entire leaderboard at once, which would be a fairly costly operation.
	if countInt > maxResultsSize {
		r = packageGenericError(400, types.LeaderboardsRangeCountInvalid, errors.New("'count' query param too large"))
		return r, nil
	}

	// Check for the existence of, and then get the value for the "pid" query parameter. Note that if it is not found,
	// we don't return - rather, we leave "pid" as an empty string so that it's ignored by the GetLeaderboards.
	var pid string
	if _, ok := request.QueryStringParameters[queryParamPublicID]; ok {
		pid = request.QueryStringParameters[queryParamPublicID]
	} else {
		pid = ""
	}

	// Attempt to get the leaderboards data for the specified range. The return value will also have an extra member
	// for the user's leaderboard row, if the pid was valid. This will be passed directly into the lambda response
	// make function, to be packaged as a JSON string in the message body.
	leaderboards, err := database.GetLeaderboards(pid, fromInt, countInt)
	if err != nil {
		r = packageLeaderboardsError(err)
		return r, nil
	}

	// Package the return payload in a lambda response.
	r = types.MakeLambdaResponse(200, types.Success, leaderboards)

	return r, nil
}
