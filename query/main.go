package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	ID string `json:"id"`
}

type MyResponse struct {
	Data json.RawMessage `json:"data"`
}

func HandleRequest(ctx context.Context, event json.RawMessage) (events.APIGatewayProxyResponse, error) {
	var myEvent MyEvent
	err := json.Unmarshal(event, &myEvent)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Error unmarshalling request", StatusCode: 400}, err
	}

	// In a real application, you would fetch data from the read model
	// based on the event.ID
	mockData := json.RawMessage(`{"id":"` + myEvent.ID + `","name":"Mock Patient"}`)

	return events.APIGatewayProxyResponse{Body: string(mockData), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
