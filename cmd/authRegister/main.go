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

type RegisterRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	Role             string `json:"role"`
	TenantID         string `json:"tenant_id"`
	NavigatorAdminID string `json:"navigator_admin_id"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{Body: "Invalid request", StatusCode: 400}, nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "AWS config error", StatusCode: 500}, nil
	}
	authService := service.NewAuthService(cfg, os.Getenv("COGNITO_USER_POOL_ID"), os.Getenv("COGNITO_CLIENT_ID"))

	_, err = authService.CreateCognitoUser(req.Email, req.Password, req.Role, req.TenantID, req.NavigatorAdminID)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Failed to create user", StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: "User registered successfully", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
