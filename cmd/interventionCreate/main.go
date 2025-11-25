package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	internalevents "github.com/lambda/internal/events"
	"github.com/lambda/internal/repository"
	"github.com/lambda/internal/service"
)

var (
	db                  *gorm.DB
	interventionService *service.InterventionService
)

func init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=write_model port=5432 sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}

	streamName := getEnv("KINESIS_STREAM_NAME", "intervention-events")
	eventPublisher := internalevents.NewKinesisEventPublisher(cfg, streamName)

	interventionRepo := repository.NewInterventionRepository(db)
	interventionService = service.NewInterventionService(interventionRepo, eventPublisher)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	tenantID := request.Headers["X-Tenant-ID"]
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	userID := request.Headers["X-User-ID"]
	if userID == "" {
		userID = "default-user"
	}

	var req service.CreateInterventionsRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid request body"}`,
		}, nil
	}

	resp, err := interventionService.CreateInterventions(ctx, tenantID, userID, &req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "` + err.Error() + `"}`,
		}, nil
	}

	body, _ := json.Marshal(resp)
	return events.APIGatewayProxyResponse{
		StatusCode: 201,
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
