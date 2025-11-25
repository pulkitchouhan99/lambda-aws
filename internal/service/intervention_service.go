package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lambda/internal/domain"
	"github.com/lambda/internal/events"
	"github.com/lambda/internal/repository"
	"github.com/lib/pq"
)

type InterventionService struct {
	repo      *repository.InterventionRepository
	publisher events.EventPublisher
}

func NewInterventionService(repo *repository.InterventionRepository, publisher events.EventPublisher) *InterventionService {
	return &InterventionService{
		repo:      repo,
		publisher: publisher,
	}
}

type InterventionItem struct {
	Type            domain.InterventionType `json:"type"`
	Title           string                  `json:"title"`
	ScheduleInDay   *time.Time              `json:"schedule_in_day,omitempty"`
	DueInDay        *time.Time              `json:"due_in_day,omitempty"`
	ReferralReasons []string                `json:"referral_reasons,omitempty"`
	Problems        []string                `json:"problems,omitempty"`
	AssignedTo      *string                 `json:"assigned_to,omitempty"`
	AssignedTeam    *string                 `json:"assigned_team,omitempty"`
	Priority        *string                 `json:"priority,omitempty"`
	Description     *string                 `json:"description,omitempty"`
}

type CreateInterventionsRequest struct {
	PatientID    string             `json:"patient_id"`
	ScreeningID  string             `json:"screening_id"`
	Items        []InterventionItem `json:"items"`
	NotifyPeople []string           `json:"notify_people,omitempty"`
	CreatedFrom  *string            `json:"created_from,omitempty"`
}

type CreatedTask struct {
	TaskID       string `json:"task_id"`
	AssigneeRole string `json:"assignee_role"`
}

type CreateInterventionsResponse struct {
	InterventionIDs []string      `json:"intervention_ids"`
	CreatedTasks    []CreatedTask `json:"created_tasks"`
}

func (s *InterventionService) CreateInterventions(ctx context.Context, tenantID, userID string, req *CreateInterventionsRequest) (*CreateInterventionsResponse, error) {
	response := &CreateInterventionsResponse{
		InterventionIDs: []string{},
		CreatedTasks:    []CreatedTask{},
	}

	for _, item := range req.Items {
		priority := "medium"
		if item.Priority != nil && *item.Priority != "" {
			priority = *item.Priority
		}

		intervention := &domain.Intervention{
			ID:              "int_" + uuid.New().String(),
			TenantID:        tenantID,
			PatientID:       req.PatientID,
			ScreeningID:     req.ScreeningID,
			Type:            item.Type,
			Title:           item.Title,
			Description:     item.Description,
			Status:          domain.StatusPending,
			Priority:        priority,
			CreatedBy:       userID,
			AssignedTo:      item.AssignedTo,
			AssignedTeam:    item.AssignedTeam,
			ReferralReasons: pq.StringArray(item.ReferralReasons),
			Problems:        pq.StringArray(item.Problems),
			DueAt:           item.DueInDay,
		}

		if err := s.repo.Create(ctx, intervention); err != nil {
			return nil, fmt.Errorf("failed to create intervention: %w", err)
		}

		response.InterventionIDs = append(response.InterventionIDs, intervention.ID)

		assigneeRole := s.getAssigneeRoleForInterventionType(item.Type)
		response.CreatedTasks = append(response.CreatedTasks, CreatedTask{
			TaskID:       "",
			AssigneeRole: assigneeRole,
		})

		if s.publisher != nil {
			event := events.NewInterventionCreatedEvent(&events.InterventionCreatedEvent{
				InterventionID:  intervention.ID,
				TenantID:        intervention.TenantID,
				PatientID:       intervention.PatientID,
				ScreeningID:     intervention.ScreeningID,
				Type:            intervention.Type,
				Title:           intervention.Title,
				Description:     intervention.Description,
				Status:          intervention.Status,
				Priority:        intervention.Priority,
				CreatedBy:       intervention.CreatedBy,
				AssignedTo:      intervention.AssignedTo,
				AssignedTeam:    intervention.AssignedTeam,
				DueAt:           intervention.DueAt,
				ReferralReasons: intervention.ReferralReasons,
				Problems:        intervention.Problems,
				CreatedAt:       intervention.CreatedAt,
			})
			if err := s.publisher.Publish(ctx, event); err != nil {
				return nil, fmt.Errorf("failed to publish intervention created event: %w", err)
			}
		}
	}

	return response, nil
}

func (s *InterventionService) GetInterventionByID(ctx context.Context, tenantID, interventionID string) (*domain.Intervention, error) {
	return s.repo.GetByID(ctx, interventionID, tenantID)
}

func (s *InterventionService) ListInterventions(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*domain.Intervention, error) {
	return s.repo.List(ctx, tenantID, filters)
}

func (s *InterventionService) UpdateIntervention(ctx context.Context, tenantID string, interventionID string, updates map[string]interface{}) error {
	intervention, err := s.repo.GetByID(ctx, interventionID, tenantID)
	if err != nil {
		return err
	}

	if assignedTo, ok := updates["assigned_to"].(string); ok {
		intervention.AssignedTo = &assignedTo
	}
	if assignedTeam, ok := updates["assigned_team"].(string); ok {
		intervention.AssignedTeam = &assignedTeam
	}
	if priority, ok := updates["priority"].(string); ok {
		intervention.Priority = priority
	}
	if notes, ok := updates["notes"].(string); ok {
		intervention.Notes = &notes
	}
	if problems, ok := updates["problems"].([]interface{}); ok {
		var problemsStr []string
		for _, p := range problems {
			if s, ok := p.(string); ok {
				problemsStr = append(problemsStr, s)
			}
		}
		intervention.Problems = pq.StringArray(problemsStr)
	}

	if err := s.repo.Update(ctx, intervention); err != nil {
		return err
	}

	if s.publisher != nil {
		event := events.NewInterventionUpdatedEvent(&events.InterventionUpdatedEvent{
			InterventionID: interventionID,
			TenantID:       tenantID,
			UpdatedFields:  updates,
			UpdatedAt:      time.Now().UTC(),
		})
		if err := s.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish intervention updated event: %w", err)
		}
	}

	return nil
}

func (s *InterventionService) CompleteIntervention(ctx context.Context, tenantID string, interventionID string, notes string) error {
	intervention, err := s.repo.GetByID(ctx, interventionID, tenantID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	intervention.Status = domain.StatusCompleted
	intervention.CompletedAt = &now
	if notes != "" {
		intervention.Notes = &notes
	}

	if err := s.repo.Update(ctx, intervention); err != nil {
		return err
	}

	if s.publisher != nil {
		var notesPtr *string
		if notes != "" {
			notesPtr = &notes
		}
		event := events.NewInterventionCompletedEvent(&events.InterventionCompletedEvent{
			InterventionID: interventionID,
			TenantID:       tenantID,
			CompletedAt:    now,
			Notes:          notesPtr,
		})
		if err := s.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish intervention completed event: %w", err)
		}
	}

	return nil
}

func (s *InterventionService) CancelIntervention(ctx context.Context, tenantID string, interventionID string, reason string) error {
	intervention, err := s.repo.GetByID(ctx, interventionID, tenantID)
	if err != nil {
		return err
	}

	intervention.Status = domain.StatusCancelled
	if reason != "" {
		intervention.Notes = &reason
	}

	if err := s.repo.Update(ctx, intervention); err != nil {
		return err
	}

	if s.publisher != nil {
		var reasonPtr *string
		if reason != "" {
			reasonPtr = &reason
		}
		event := events.NewInterventionCancelledEvent(&events.InterventionCancelledEvent{
			InterventionID: interventionID,
			TenantID:       tenantID,
			CancelledAt:    time.Now().UTC(),
			Reason:         reasonPtr,
		})
		if err := s.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish intervention cancelled event: %w", err)
		}
	}

	return nil
}

func (s *InterventionService) GetBarrierCounts(ctx context.Context, tenantID string, filters map[string]interface{}) (*domain.BarrierResponse, error) {
	return s.repo.GetBarrierCounts(ctx, tenantID, filters)
}

func (s *InterventionService) getAssigneeRoleForInterventionType(interventionType domain.InterventionType) string {
	switch interventionType {
	case domain.TypeFinancialCounselor:
		return "financial_assist"
	case domain.TypeSocialWork:
		return "social_worker"
	case domain.TypeRegisteredDietitian:
		return "dietitian"
	case domain.TypeSpiritualCare:
		return "spiritual_care"
	case domain.TypeGeneticCounselor:
		return "genetic_counselor"
	case domain.TypeRehabilitation:
		return "rehabilitation"
	case domain.TypeClinicalTrial:
		return "clinical_trial"
	case domain.TypeTranslationServices:
		return "translator"
	case domain.TypePalliativeCare:
		return "palliative_care"
	case domain.TypeHospice:
		return "hospice"
	case domain.TypeSurvivorship:
		return "survivorship"
	case domain.TypeCommunityResource:
		return "community_resource"
	case domain.TypeCoordinationOfCare:
		return "care_coordinator"
	case domain.TypePhysicianOfficeStaff:
		return "physician_staff"
	case domain.TypeOther:
		return "navigator"
	default:
		return "navigator"
	}
}
