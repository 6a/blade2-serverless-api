package main

import (
	"encoding/json"
	"log"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/aws/aws-lambda-go/events"
)

func main() {
	database.Init()

	ev := events.APIGatewayProxyRequest{}
	// ev.Headers = make(map[string]string)
	ev.PathParameters = make(map[string]string, 1)
	ev.PathParameters["pid"] = "bqnf9bu4h65c72kc033g"
	// ev.QueryStringParameters = make(map[string]string, 2)
	// ev.QueryStringParameters["from"] = "0"
	// ev.QueryStringParameters["count"] = "10"
	// ev.QueryStringParameters["pid"] = "bqnfd6e4h65c72kc0340"

	r, err := routes.GetMatchHistory(nil, ev)

	if err != nil {
		log.Fatal(err)
	}

	data, err := json.MarshalIndent(r.Body, "", "  ")
	if err != nil {
		log.Print(r.Body)
	} else {
		log.Print(string(data))
	}
}
