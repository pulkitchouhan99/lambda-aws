package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/google/uuid"
)

type EventPublisher interface {
	Publish(ctx context.Context, event *DomainEvent) error
}

type KinesisEventPublisher struct {
	client     *kinesis.Client
	streamName string
}

func NewKinesisEventPublisher(cfg aws.Config, streamName string) *KinesisEventPublisher {
	return &KinesisEventPublisher{
		client:     kinesis.NewFromConfig(cfg),
		streamName: streamName,
	}
}

func (p *KinesisEventPublisher) Publish(ctx context.Context, event *DomainEvent) error {
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = p.client.PutRecord(ctx, &kinesis.PutRecordInput{
		StreamName:   aws.String(p.streamName),
		Data:         data,
		PartitionKey: aws.String(event.AggregateID),
	})
	if err != nil {
		return fmt.Errorf("failed to publish event to Kinesis: %w", err)
	}

	log.Printf("Published event: %s for aggregate: %s", event.EventType, event.AggregateID)
	return nil
}

func NewInterventionCreatedEvent(intervention *InterventionCreatedEvent) *DomainEvent {
	payload := map[string]interface{}{
		"intervention_id":  intervention.InterventionID,
		"tenant_id":        intervention.TenantID,
		"patient_id":       intervention.PatientID,
		"screening_id":     intervention.ScreeningID,
		"type":             intervention.Type,
		"title":            intervention.Title,
		"description":      intervention.Description,
		"status":           intervention.Status,
		"priority":         intervention.Priority,
		"created_by":       intervention.CreatedBy,
		"assigned_to":      intervention.AssignedTo,
		"assigned_team":    intervention.AssignedTeam,
		"due_at":           intervention.DueAt,
		"referral_reasons": intervention.ReferralReasons,
		"problems":         intervention.Problems,
		"created_at":       intervention.CreatedAt,
	}

	return &DomainEvent{
		EventID:     uuid.New().String(),
		EventType:   InterventionCreated,
		AggregateID: intervention.InterventionID,
		TenantID:    intervention.TenantID,
		Timestamp:   time.Now().UTC(),
		Payload:     payload,
		Metadata: map[string]string{
			"source": "intervention-service",
		},
	}
}

func NewInterventionUpdatedEvent(updated *InterventionUpdatedEvent) *DomainEvent {
	payload := map[string]interface{}{
		"intervention_id": updated.InterventionID,
		"tenant_id":       updated.TenantID,
		"updated_fields":  updated.UpdatedFields,
		"updated_at":      updated.UpdatedAt,
	}

	return &DomainEvent{
		EventID:     uuid.New().String(),
		EventType:   InterventionUpdated,
		AggregateID: updated.InterventionID,
		TenantID:    updated.TenantID,
		Timestamp:   time.Now().UTC(),
		Payload:     payload,
		Metadata: map[string]string{
			"source": "intervention-service",
		},
	}
}

func NewInterventionCompletedEvent(completed *InterventionCompletedEvent) *DomainEvent {
	payload := map[string]interface{}{
		"intervention_id": completed.InterventionID,
		"tenant_id":       completed.TenantID,
		"completed_at":    completed.CompletedAt,
		"notes":           completed.Notes,
	}

	return &DomainEvent{
		EventID:     uuid.New().String(),
		EventType:   InterventionCompleted,
		AggregateID: completed.InterventionID,
		TenantID:    completed.TenantID,
		Timestamp:   time.Now().UTC(),
		Payload:     payload,
		Metadata: map[string]string{
			"source": "intervention-service",
		},
	}
}

func NewInterventionCancelledEvent(cancelled *InterventionCancelledEvent) *DomainEvent {
	payload := map[string]interface{}{
		"intervention_id": cancelled.InterventionID,
		"tenant_id":       cancelled.TenantID,
		"cancelled_at":    cancelled.CancelledAt,
		"reason":          cancelled.Reason,
	}

	return &DomainEvent{
		EventID:     uuid.New().String(),
		EventType:   InterventionCancelled,
		AggregateID: cancelled.InterventionID,
		TenantID:    cancelled.TenantID,
		Timestamp:   time.Now().UTC(),
		Payload:     payload,
		Metadata: map[string]string{
			"source": "intervention-service",
		},
	}
}
