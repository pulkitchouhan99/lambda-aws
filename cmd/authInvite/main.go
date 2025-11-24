package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lambda/internal/domain"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "DB connection error", StatusCode: 500}, nil
	}

	invitationRepo := repository.NewInvitationRepository(db)
	invitationService := service.NewInvitationService(invitationRepo, os.Getenv("JWT_SECRET"))

	var req InviteRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{Body: "Invalid request", StatusCode: 400}, nil
	}

	// In a real application, you'd get the inviter's ID from the JWT
	invitation := &domain.Invitation{
		Email: req.Email,
		Role:  domain.Role(req.Role),
	}

	token, err := invitationService.CreateInvitation(invitation)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Failed to create invitation", StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: `{"token":"` + token + `"}`, StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
