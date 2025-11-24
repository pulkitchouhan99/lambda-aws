package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lambda/internal/domain"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

type CognitoEvent struct {
	Request Request `json:"request"`
}

type Request struct {
	UserAttributes map[string]string `json:"userAttributes"`
}

func HandleRequest(ctx context.Context, event CognitoEvent) error {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
		return err
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	// In a real application, you would parse the custom attributes and create the user
	// For now, this is a placeholder
	user := &domain.User{
		Email: event.Request.UserAttributes["email"],
	}

	if err := userService.CreateUser(user); err != nil {
		log.Printf("failed to create user in database: %v", err)
		return err
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
