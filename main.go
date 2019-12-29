package main

import (
	"fmt"
	"log"

	"github.com/6a/blade-ii-api/internal/database"
	"github.com/6a/blade-ii-api/internal/routes"
	"github.com/aws/aws-lambda-go/events"
)

func main() {
	database.Init()

	ev := events.APIGatewayProxyRequest{}
	ev.Headers = make(map[string]string)

	handle := "ケビン・グラハム17"
	email := "test37@gmail.com"
	pass := "testpaa1"

	ev.Body = fmt.Sprintf(`{"handle": "%v", "email": "%v", "password": "%v"}`, handle, email, pass)
	r, err := routes.CreateAccount(nil, ev)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(r)
}
