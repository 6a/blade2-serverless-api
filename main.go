package main

import (
	"log"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/aws/aws-lambda-go/events"
)

func main() {
	database.Init()

	ev := events.APIGatewayProxyRequest{}
	ev.Headers = make(map[string]string)
	// ev.PathParameters = make(map[string]string, 1)
	// ev.PathParameters["pid"] = ""

	r, err := routes.GetProfile(nil, ev)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(r)
}
