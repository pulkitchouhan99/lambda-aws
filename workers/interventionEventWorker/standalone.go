//go:build standalone
// +build standalone

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var readDB *gorm.DB
var sqsClient *sqs.Client
var queueURL string

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
	readDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to read database: %v", err)
	}
	log.Println("Connected to Read DB successfully")

	ctx := context.Background()
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		localstackURL := getEnv("LOCALSTACK_URL", "http://localhost:4566")
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           localstackURL,
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		})),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	sqsClient = sqs.NewFromConfig(cfg)
	queueURL = getEnv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/intervention-events")
	log.Printf("Connected to SQS queue: %s", queueURL)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down worker...")
		cancel()
	}()

	log.Println("Worker started. Polling SQS for messages...")
	pollMessages(ctx)
}

func pollMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
			receiveMessages(ctx)
			time.Sleep(2 * time.Second)
		}
	}
}

func receiveMessages(ctx context.Context) {
	result, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5,
		VisibilityTimeout:   30,
	})
	if err != nil {
		log.Printf("Error receiving messages: %v", err)
		return
	}

	for _, message := range result.Messages {
		if err := processMessage(ctx, message); err != nil {
			log.Printf("Error processing message: %v", err)
		} else {
			_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: message.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Error deleting message: %v", err)
			}
		}
	}
}

func processMessage(ctx context.Context, message types.Message) error {
	log.Printf("Processing message: %s", *message.MessageId)

	var domainEvent map[string]interface{}
	if err := json.Unmarshal([]byte(*message.Body), &domainEvent); err != nil {
		return err
	}

	if err := processEvent(ctx, domainEvent); err != nil {
		return err
	}

	log.Printf("Successfully processed event: %s", domainEvent["event_id"])
	return nil
}

func processEvent(ctx context.Context, event map[string]interface{}) error {
	eventType := event["event_type"].(string)
	
	switch eventType {
	case "intervention.created":
		return handleInterventionCreated(ctx, event)
	case "intervention.updated":
		return handleInterventionUpdated(ctx, event)
	case "intervention.completed":
		return handleInterventionCompleted(ctx, event)
	case "intervention.cancelled":
		return handleInterventionCancelled(ctx, event)
	default:
		log.Printf("Unknown event type: %s", eventType)
		return nil
	}
}

func handleInterventionCreated(ctx context.Context, event map[string]interface{}) error {
	payload := event["payload"].(map[string]interface{})

	// Handle array types properly
	var referralReasons []string
	if rr, ok := payload["referral_reasons"].([]interface{}); ok {
		for _, v := range rr {
			if s, ok := v.(string); ok {
				referralReasons = append(referralReasons, s)
			}
		}
	} else if rr, ok := payload["referral_reasons"].([]string); ok {
		referralReasons = rr
	}

	var problems []string
	if p, ok := payload["problems"].([]interface{}); ok {
		for _, v := range p {
			if s, ok := v.(string); ok {
				problems = append(problems, s)
			}
		}
	} else if p, ok := payload["problems"].([]string); ok {
		problems = p
	}

	// Convert arrays to PostgreSQL array format
	referralReasonsStr := "{}"
	if len(referralReasons) > 0 {
		quotedReasons := make([]string, len(referralReasons))
		for i, reason := range referralReasons {
			quotedReasons[i] = "'" + strings.ReplaceAll(reason, "'", "''") + "'"
		}
		referralReasonsStr = "{" + strings.Join(quotedReasons, ",") + "}"
	}

	problemsStr := "{}"
	if len(problems) > 0 {
		quotedProblems := make([]string, len(problems))
		for i, problem := range problems {
			quotedProblems[i] = "'" + strings.ReplaceAll(problem, "'", "''") + "'"
		}
		problemsStr = "{" + strings.Join(quotedProblems, ",") + "}"
	}

	// Use raw SQL to insert with proper array handling
	query := `INSERT INTO interventions_projection 
		(id, tenant_id, patient_id, screening_id, type, title, description, status, priority, 
		 created_by, assigned_to, assigned_team, due_at, referral_reasons, problems, 
		 created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::text[], ?::text[], ?, ?)`

	result := readDB.Exec(query,
		getString(payload["intervention_id"]),
		getString(payload["tenant_id"]),
		getString(payload["patient_id"]),
		getString(payload["screening_id"]),
		getString(payload["type"]),
		getString(payload["title"]),
		getStringPtr(payload["description"]),
		getString(payload["status"]),
		getString(payload["priority"]),
		getString(payload["created_by"]),
		getStringPtr(payload["assigned_to"]),
		getStringPtr(payload["assigned_team"]),
		getStringPtr(payload["due_at"]),
		referralReasonsStr,
		problemsStr,
		getString(payload["created_at"]),
		getString(payload["created_at"]),
	)

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Created intervention projection: %s", payload["intervention_id"])
	return nil
}

func handleInterventionUpdated(ctx context.Context, event map[string]interface{}) error {
	payload := event["payload"].(map[string]interface{})
	interventionID := getString(payload["intervention_id"])
	
	updatedFields := make(map[string]interface{})
	if uf, ok := payload["updated_fields"].(map[string]interface{}); ok {
		updatedFields = uf
	}
	
	updatedFields["updated_at"] = getString(payload["updated_at"])

	// Build update query dynamically
	setParts := []string{}
	values := []interface{}{}
	
	for key, value := range updatedFields {
		setParts = append(setParts, key+" = ?")
		values = append(values, value)
	}
	
	if len(setParts) == 0 {
		return nil
	}
	
	query := "UPDATE interventions_projection SET " + strings.Join(setParts, ", ") + 
		" WHERE id = ? AND tenant_id = ?"
	values = append(values, interventionID, event["tenant_id"])

	result := readDB.Exec(query, values...)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Updated intervention projection: %s", interventionID)
	return nil
}

func handleInterventionCompleted(ctx context.Context, event map[string]interface{}) error {
	payload := event["payload"].(map[string]interface{})
	interventionID := getString(payload["intervention_id"])

	updates := map[string]interface{}{
		"status":       "completed",
		"completed_at": getString(payload["completed_at"]),
		"updated_at":   getString(payload["completed_at"]),
	}
	if notes := payload["notes"]; notes != nil {
		if noteStr, ok := notes.(string); ok {
			updates["notes"] = &noteStr
		}
	}

	// Build update query
	setParts := []string{}
	values := []interface{}{}
	
	for key, value := range updates {
		setParts = append(setParts, key+" = ?")
		values = append(values, value)
	}
	
	query := "UPDATE interventions_projection SET " + strings.Join(setParts, ", ") + 
		" WHERE id = ? AND tenant_id = ?"
	values = append(values, interventionID, event["tenant_id"])

	result := readDB.Exec(query, values...)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Completed intervention projection: %s", interventionID)
	return nil
}

func handleInterventionCancelled(ctx context.Context, event map[string]interface{}) error {
	payload := event["payload"].(map[string]interface{})
	interventionID := getString(payload["intervention_id"])

	updates := map[string]interface{}{
		"status":     "cancelled",
		"updated_at": getString(payload["cancelled_at"]),
	}
	if reason := payload["reason"]; reason != nil {
		if reasonStr, ok := reason.(string); ok {
			updates["notes"] = &reasonStr
		}
	}

	// Build update query
	setParts := []string{}
	values := []interface{}{}
	
	for key, value := range updates {
		setParts = append(setParts, key+" = ?")
		values = append(values, value)
	}
	
	query := "UPDATE interventions_projection SET " + strings.Join(setParts, ", ") + 
		" WHERE id = ? AND tenant_id = ?"
	values = append(values, interventionID, event["tenant_id"])

	result := readDB.Exec(query, values...)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Cancelled intervention projection: %s", interventionID)
	return nil
}

// Helper functions
func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return &s
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}