package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

type CognitoEvent struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	UserAttributes map[string]string `json:"userAttributes"`
}

type Response struct {
	AutoConfirmUser bool `json:"autoConfirmUser"`
}

func HandleRequest(ctx context.Context, event CognitoEvent) (CognitoEvent, error) {
	event.Response.AutoConfirmUser = true
	// In a real application, you might add more logic here,
	// like validating the invitation token again.
	return event, nil
}

func main() {
	lambda.Start(HandleRequest)
}
