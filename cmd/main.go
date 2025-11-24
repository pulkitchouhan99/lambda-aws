package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Name string `json:"name"`
}

type MyResponse struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	fmt.Println("invoked")
	var myEvent MyEvent
	err := json.Unmarshal(event, &myEvent)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error unmarshalling request", StatusCode: 400}, err
	}

	return events.APIGatewayProxyResponse{Body: fmt.Sprintf("Hello, %s!", myEvent.Name), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
