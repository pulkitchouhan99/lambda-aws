package events

import (
	"time"

	"github.com/lambda/internal/domain"
)

type EventType string

const (
	InterventionCreated   EventType = "intervention.created"
	InterventionUpdated   EventType = "intervention.updated"
	InterventionCompleted EventType = "intervention.completed"
	InterventionCancelled EventType = "intervention.cancelled"
)

type DomainEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   EventType              `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    string                 `json:"tenant_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Payload     map[string]interface{} `json:"payload"`
	Metadata    map[string]string      `json:"metadata"`
}

type InterventionCreatedEvent struct {
	InterventionID  string                    `json:"intervention_id"`
	TenantID        string                    `json:"tenant_id"`
	PatientID       string                    `json:"patient_id"`
	ScreeningID     string                    `json:"screening_id"`
	Type            domain.InterventionType   `json:"type"`
	Title           string                    `json:"title"`
	Description     *string                   `json:"description,omitempty"`
	Status          domain.InterventionStatus `json:"status"`
	Priority        string                    `json:"priority"`
	CreatedBy       string                    `json:"created_by"`
	AssignedTo      *string                   `json:"assigned_to,omitempty"`
	AssignedTeam    *string                   `json:"assigned_team,omitempty"`
	DueAt           *time.Time                `json:"due_at,omitempty"`
	ReferralReasons []string                  `json:"referral_reasons"`
	Problems        []string                  `json:"problems"`
	CreatedAt       time.Time                 `json:"created_at"`
}

type InterventionUpdatedEvent struct {
	InterventionID string                 `json:"intervention_id"`
	TenantID       string                 `json:"tenant_id"`
	UpdatedFields  map[string]interface{} `json:"updated_fields"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type InterventionCompletedEvent struct {
	InterventionID string     `json:"intervention_id"`
	TenantID       string     `json:"tenant_id"`
	CompletedAt    time.Time  `json:"completed_at"`
	Notes          *string    `json:"notes,omitempty"`
}

type InterventionCancelledEvent struct {
	InterventionID string     `json:"intervention_id"`
	TenantID       string     `json:"tenant_id"`
	CancelledAt    time.Time  `json:"cancelled_at"`
	Reason         *string    `json:"reason,omitempty"`
}
