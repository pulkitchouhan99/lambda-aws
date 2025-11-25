package repository

import (
	"context"
	"gorm.io/gorm"
)

type InterventionProjection struct {
	ID              string   `gorm:"primaryKey;type:text" json:"id"`
	TenantID        string   `gorm:"type:text;index" json:"tenant_id"`
	PatientID       string   `gorm:"type:text;index" json:"patient_id"`
	ScreeningID     string   `gorm:"type:text;not null" json:"screening_id"`
	Type            string   `gorm:"type:text;not null" json:"type"`
	Title           string   `gorm:"type:text;not null" json:"title"`
	Description     *string  `gorm:"type:text" json:"description,omitempty"`
	Status          string   `gorm:"type:text;not null" json:"status"`
	Priority        string   `gorm:"type:text" json:"priority"`
	CreatedBy       string   `gorm:"type:text;not null" json:"created_by"`
	AssignedTo      *string  `gorm:"type:text" json:"assigned_to,omitempty"`
	AssignedTeam    *string  `gorm:"type:text" json:"assigned_team,omitempty"`
	DueAt           *string  `gorm:"type:timestamptz" json:"due_at,omitempty"`
	CompletedAt     *string  `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	LinkedTaskID    *string  `gorm:"type:text" json:"linked_task_id,omitempty"`
	ReferralReasons []string `gorm:"type:text[]" json:"referral_reasons"`
	Problems        []string `gorm:"type:text[]" json:"problems"`
	Notes           *string  `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       string   `gorm:"type:timestamptz" json:"created_at"`
	UpdatedAt       string   `gorm:"type:timestamptz" json:"updated_at"`
}

func (InterventionProjection) TableName() string {
	return "interventions_projection"
}

type InterventionProjectionRepository struct {
	db *gorm.DB
}

func NewInterventionProjectionRepository(db *gorm.DB) *InterventionProjectionRepository {
	return &InterventionProjectionRepository{db: db}
}

func (r *InterventionProjectionRepository) GetByID(ctx context.Context, id, tenantID string) (*InterventionProjection, error) {
	var intervention InterventionProjection
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&intervention).Error
	if err != nil {
		return nil, err
	}
	return &intervention, nil
}

func (r *InterventionProjectionRepository) List(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*InterventionProjection, error) {
	var interventions []*InterventionProjection

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if interventionType, ok := filters["type"].(string); ok && interventionType != "" {
		query = query.Where("type = ?", interventionType)
	}
	if patientID, ok := filters["patient_id"].(string); ok && patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if screeningID, ok := filters["screening_id"].(string); ok && screeningID != "" {
		query = query.Where("screening_id = ?", screeningID)
	}
	if screeningIDs, ok := filters["screening_ids"].([]string); ok && len(screeningIDs) > 0 {
		query = query.Where("screening_id IN ?", screeningIDs)
	}
	if createdBy, ok := filters["created_by"].(string); ok && createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if createdByIDs, ok := filters["created_by_ids"].([]string); ok && len(createdByIDs) > 0 {
		query = query.Where("created_by IN ?", createdByIDs)
	}
	if assignedTeam, ok := filters["assigned_team"].(string); ok && assignedTeam != "" {
		query = query.Where("assigned_team = ?", assignedTeam)
	}

	err := query.Order("created_at DESC").Find(&interventions).Error
	if err != nil {
		return nil, err
	}

	return interventions, nil
}

func (r *InterventionProjectionRepository) Count(ctx context.Context, tenantID string, filters map[string]interface{}) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&InterventionProjection{}).Where("tenant_id = ?", tenantID)

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if interventionType, ok := filters["type"].(string); ok && interventionType != "" {
		query = query.Where("type = ?", interventionType)
	}
	if patientID, ok := filters["patient_id"].(string); ok && patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	err := query.Count(&count).Error
	return count, err
}
