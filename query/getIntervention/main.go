package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/lambda/internal/repository"
)

var (
	db             *gorm.DB
	projectionRepo *repository.InterventionProjectionRepository
)

func init() {
	dsn := os.Getenv("READ_DB_URL")
	if dsn == "" {
		host := getEnv("READ_DB_HOST", "localhost")
		port := getEnv("READ_DB_PORT", "5433")
		user := getEnv("READ_DB_USER", "postgres")
		password := getEnv("READ_DB_PASSWORD", "postgres")
		dbname := getEnv("READ_DB_NAME", "read_model")
		sslmode := getEnv("READ_DB_SSLMODE", "disable")
		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to read database: " + err.Error())
	}

	projectionRepo = repository.NewInterventionProjectionRepository(db)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tenantID := request.Headers["X-Tenant-ID"]
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	interventionID := request.PathParameters["id"]
	if interventionID == "" {
		interventionID = request.QueryStringParameters["id"]
	}

	if interventionID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Missing intervention ID"}`,
		}, nil
	}

	intervention, err := projectionRepo.GetByID(ctx, interventionID, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       `{"error": "Intervention not found"}`,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "` + err.Error() + `"}`,
		}, nil
	}

	body, _ := json.Marshal(intervention)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	lambda.Start(HandleRequest)
}
