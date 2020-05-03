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
// extra blob containing the row for the user specified by the (pid) public ID query param. If pid is unspecified,
// the user blob is returned with default values
func GetLeaderboards(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	var from string
	if _, ok := request.QueryStringParameters[queryParamFrom]; ok {
		from = request.QueryStringParameters[queryParamFrom]
	} else {
		r = packageGenericError(400, types.LeaderboardsRangeFromMissing, errors.New("'from' query param missing"))
		return r, nil
	}

	fromInt, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		r = packageGenericError(400, types.LeaderboardsRangeFromInvalid, errors.New("'from' query param invalid"))
		return r, nil
	}

	var count string
	if _, ok := request.QueryStringParameters[queryParamCount]; ok {
		count = request.QueryStringParameters[queryParamCount]
	} else {
		r = packageGenericError(400, types.LeaderboardsRangeCountMissing, errors.New("'count' query param missing"))
		return r, nil
	}

	countInt, err := strconv.ParseUint(count, 10, 64)
	if err != nil {
		r = packageGenericError(400, types.LeaderboardsRangeCountInvalid, errors.New("'count' query param invalid"))
		return r, nil
	}

	if countInt > maxResultsSize {
		r = packageGenericError(400, types.LeaderboardsRangeCountInvalid, errors.New("'count' query param too large"))
		return r, nil
	}

	var pid string
	if _, ok := request.QueryStringParameters[queryParamPublicID]; ok {
		pid = request.QueryStringParameters[queryParamPublicID]
	} else {
		pid = ""
	}

	leaderboards, err := database.GetLeaderboards(pid, fromInt, countInt)
	if err != nil {
		r = packageLeaderboardsError(err)
		return r, nil
	}

	r = types.MakeLambdaResponse(200, types.Success, leaderboards)

	return r, nil
}
