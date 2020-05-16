// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package main implements interfaces for the lambda functions to be run, either through AWS
// lambdas, or running locally.
package main

import (
	"log"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/aws/aws-lambda-go/events"
)

// This version of the main function is for running locally.
func main() {

	// Initialize the database package.
	database.Init()

	// Create a new empty AWS API gateway proxy request.
	ev := events.APIGatewayProxyRequest{}
	ev.Path = "leaderboards"
	ev.QueryStringParameters = make(map[string]string, 0)
	ev.QueryStringParameters["from"] = "0"
	ev.QueryStringParameters["count"] = "100"
	ev.QueryStringParameters["pid"] = "bqnfdcu4h65c72kc037g"

	// Perform a test call to one of the routes.
	r, err := routes.GetLeaderboards(nil, ev)

	// Check for an error.
	if err != nil {
		log.Fatal(err)
	}

	// Print the returned result on success.
	log.Print(r)
}
