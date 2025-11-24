package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"os"

	"github.com/lambda/internal/service"
)

type VerifyOtpRequest struct {
	Email   string `json:"email"`
	OTP     string `json:"otp"`
	Session string `json:"session"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req VerifyOtpRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{Body: "Invalid request", StatusCode: 400}, nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "AWS config error", StatusCode: 500}, nil
	}
	authService := service.NewAuthService(cfg, os.Getenv("COGNito_USER_POOL_ID"), os.Getenv("COGNITO_CLIENT_ID"))

	authResult, err := authService.VerifyOTPChallenge(req.Email, req.OTP, req.Session)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Failed to verify OTP", StatusCode: 500}, nil
	}

	body, _ := json.Marshal(authResult)
	return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
