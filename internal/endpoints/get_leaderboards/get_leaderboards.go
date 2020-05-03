package main

import (
	"context"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func functionWrapper(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	return routes.GetLeaderboards(ctx, request)
}

func main() {
	database.Init()
	lambda.Start(functionWrapper)
}
