package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"os"

	    "github.com/lambda/internal/service")

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	invitationService := service.NewInvitationService(nil, os.Getenv("JWT_SECRET"))

	tokenStr, ok := request.QueryStringParameters["token"]
	if !ok {
		return events.APIGatewayProxyResponse{Body: "Missing token", StatusCode: 400}, nil
	}

	token, err := invitationService.ValidateInvitationToken(tokenStr)
	if err != nil || !token.Valid {
		return events.APIGatewayProxyResponse{Body: "Invalid token", StatusCode: 400}, nil
	}

	return events.APIGatewayProxyResponse{Body: "Token is valid", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
