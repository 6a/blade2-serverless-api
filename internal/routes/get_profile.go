package routes

import (
	"context"
	"errors"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
)

const publicIDParameterKey = "pid"

// GetProfile returns the profile data for the user specified by the public ID (in the url)
func GetProfile(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	var pid string
	if _, ok := request.PathParameters[publicIDParameterKey]; ok {
		pid = request.PathParameters[publicIDParameterKey]
	} else {
		r = packageGenericError(400, types.ProfileGetPublicIDMising, errors.New("Public ID parameter missing"))
		return r, nil
	}

	DBID, err := database.GetDBID(pid)
	if err != nil {
		r = packageGenericError(404, types.ProfileGetPublicIDNotFound, errors.New("No matching profile found"))
		return r, nil
	}

	profile, err := database.GetProfile(DBID)
	if err != nil {
		r = packageProfileGetError(pid, err)
		return r, nil
	}

	r = types.MakeLambdaResponse(200, types.Success, profile)

	return r, nil
}
