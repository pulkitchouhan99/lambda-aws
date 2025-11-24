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
	"github.com/lambda/internal/service"
)

var (
	db                  *gorm.DB
	interventionService *service.InterventionService
)

func init() {
	// Initialize database connection once during cold start
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=write_model port=5432 sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	// Initialize repositories and services
	interventionRepo := repository.NewInterventionRepository(db)
	interventionService = service.NewInterventionService(interventionRepo)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get tenant ID from headers or authorizer context
	tenantID := request.Headers["X-Tenant-ID"]
	if tenantID == "" {
		tenantID = "default-tenant"
	}

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if status := request.QueryStringParameters["status"]; status != "" {
		filters["status"] = status
	}
	if interventionType := request.QueryStringParameters["type"]; interventionType != "" {
		filters["type"] = interventionType
	}
	if patientID := request.QueryStringParameters["patient_id"]; patientID != "" {
		filters["patient_id"] = patientID
	}
	if screeningID := request.QueryStringParameters["screening_id"]; screeningID != "" {
		filters["screening_id"] = screeningID
	}

	interventions, err := interventionService.ListInterventions(ctx, tenantID, filters)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "` + err.Error() + `"}`,
		}, nil
	}

	body, _ := json.Marshal(map[string]interface{}{
		"interventions": interventions,
		"total":         len(interventions),
	})

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
