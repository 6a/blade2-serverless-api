// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package main implements interfaces for the lambda functions to be run, either through AWS
// lambdas, or running locally.
package main

import (
	"context"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/6a/blade-ii-api/internal/types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// functionWrapper is used so that this file can easily be copied and converted for use with another route.
func functionWrapper(ctx context.Context, request events.APIGatewayProxyRequest) (r types.LambdaResponse, err error) {
	return routes.GetMatchHistory(ctx, request)
}

func main() {

	// Initialize the database package.
	database.Init()

	// Start the lambda function handler.
	lambda.Start(functionWrapper)
}
