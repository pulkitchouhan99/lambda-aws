package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/lambda/internal/domain"
	"gorm.io/gorm"
)

type InterventionRepository struct {
	db *gorm.DB
}

func NewInterventionRepository(db *gorm.DB) *InterventionRepository {
	return &InterventionRepository{
		db: db,
	}
}

func (r *InterventionRepository) Create(ctx context.Context, intervention *domain.Intervention) error {
	return r.db.WithContext(ctx).Create(intervention).Error
}

func (r *InterventionRepository) GetByID(ctx context.Context, id string, tenantID string) (*domain.Intervention, error) {
	var intervention domain.Intervention
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("User").
		First(&intervention).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("intervention not found")
		}
		return nil, err
	}
	return &intervention, nil
}

func (r *InterventionRepository) GetByPatientID(ctx context.Context, patientID string, tenantID string) ([]*domain.Intervention, error) {
	var interventions []*domain.Intervention
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND tenant_id = ?", patientID, tenantID).
		Preload("User").
		Order("created_at DESC").
		Find(&interventions).Error
	return interventions, err
}

func (r *InterventionRepository) GetByScreeningID(ctx context.Context, screeningID string, tenantID string) ([]*domain.Intervention, error) {
	var interventions []*domain.Intervention
	err := r.db.WithContext(ctx).
		Where("screening_id = ? AND tenant_id = ?", screeningID, tenantID).
		Preload("User").
		Order("created_at DESC").
		Find(&interventions).Error
	return interventions, err
}

func (r *InterventionRepository) Update(ctx context.Context, intervention *domain.Intervention) error {
	return r.db.WithContext(ctx).Save(intervention).Error
}

func (r *InterventionRepository) Delete(ctx context.Context, id string, tenantID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&domain.Intervention{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("intervention not found")
	}
	return nil
}

func (r *InterventionRepository) List(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*domain.Intervention, error) {
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if interventionType, ok := filters["type"]; ok {
		query = query.Where("type = ?", interventionType)
	}
	if assignedTeam, ok := filters["assigned_team"]; ok {
		query = query.Where("assigned_team = ?", assignedTeam)
	}
	if patientID, ok := filters["patient_id"]; ok {
		query = query.Where("patient_id = ?", patientID)
	}
	if screeningID, ok := filters["screening_id"]; ok {
		query = query.Where("screening_id = ?", screeningID)
	}
	if screeningIDs, ok := filters["screening_ids"].([]string); ok && len(screeningIDs) > 0 {
		query = query.Where("screening_id IN (?)", screeningIDs)
	}
	if createdByIDs, ok := filters["created_by_ids"].([]string); ok && len(createdByIDs) > 0 {
		query = query.Where("created_by IN (?)", createdByIDs)
	} else if createdBy, ok := filters["created_by"]; ok {
		query = query.Where("created_by = ?", createdBy)
	}

	query = query.Preload("User")
	var interventions []*domain.Intervention
	err := query.Order("created_at DESC").Find(&interventions).Error
	return interventions, err
}

func (r *InterventionRepository) GetBarrierCounts(ctx context.Context, tenantID string, filters map[string]interface{}) (*domain.BarrierResponse, error) {
	// Simplified mock implementation - in production this would match the blueprint implementation
	return &domain.BarrierResponse{
		ChartData:   []*domain.BarrierCount{},
		SubtypeData: []*domain.BarrierSubtype{},
	}, nil
}
