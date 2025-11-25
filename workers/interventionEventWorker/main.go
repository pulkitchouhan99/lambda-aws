//go:build !standalone
// +build !standalone

package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"gorm.io/gorm"
)

var readDB *gorm.DB

func init() {
	// This would be initialized in the Lambda environment
	// For now, we'll leave it as nil since the Lambda handler
	// will receive the DB connection through the context or
	// environment variables
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		log.Printf("Processing message: %s", record.MessageId)

		var domainEvent map[string]interface{}
		if err := json.Unmarshal([]byte(record.Body), &domainEvent); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		if err := processEvent(ctx, domainEvent); err != nil {
			log.Printf("Failed to process event %s: %v", domainEvent["event_id"], err)
			return err
		}

		log.Printf("Successfully processed event: %s (type: %s)", domainEvent["event_id"], domainEvent["event_type"])
	}

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
	// This is a simplified version for Lambda that would typically
	// call a service layer function
	log.Printf("Handling intervention created event: %s", event["event_id"])
	return nil
}

func handleInterventionUpdated(ctx context.Context, event map[string]interface{}) error {
	// This is a simplified version for Lambda that would typically
	// call a service layer function
	log.Printf("Handling intervention updated event: %s", event["event_id"])
	return nil
}

func handleInterventionCompleted(ctx context.Context, event map[string]interface{}) error {
	// This is a simplified version for Lambda that would typically
	// call a service layer function
	log.Printf("Handling intervention completed event: %s", event["event_id"])
	return nil
}

func handleInterventionCancelled(ctx context.Context, event map[string]interface{}) error {
	// This is a simplified version for Lambda that would typically
	// call a service layer function
	log.Printf("Handling intervention cancelled event: %s", event["event_id"])
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

func main() {
	lambda.Start(HandleRequest)
}