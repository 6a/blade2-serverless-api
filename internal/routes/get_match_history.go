package routes

import (
	"context"
	"errors"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

// GetMatchHistory returns match history for the user specified by public id in the path /matches/{publicID}
func GetMatchHistory(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	var pid string
	if _, ok := request.PathParameters[publicIDParameterKey]; ok {
		pid = request.PathParameters[publicIDParameterKey]
	} else {
		r = packageGenericError(400, types.MatchHistoryGetPublicIDMising, errors.New("Public ID parameter missing"))
		return r, nil
	}

	DBID, err := database.GetDBID(pid)
	if err != nil {
		r = packageGenericError(404, types.MatchHistoryGetPublicIDNotFound, errors.New("Public ID not found"))
		return r, nil
	}

	matchHistory, err := database.GetMatchHistory(DBID)
	if err != nil {
		r = packageMatchHistoryError(err)
		return r, nil
	}

	r = types.MakeLambdaResponse(200, types.Success, matchHistory)

	return r, nil
}
