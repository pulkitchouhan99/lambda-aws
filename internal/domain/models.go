package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Patient struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Screening struct {
	ID        uuid.UUID `json:"id"`
	PatientID uuid.UUID `json:"patient_id"`
	Answers   []string  `json:"answers"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InterventionType string

const (
	TypeFinancialCounselor   InterventionType = "financial_counselor"
	TypeSocialWork           InterventionType = "social_work"
	TypeRegisteredDietitian  InterventionType = "registered_dietitian"
	TypeSpiritualCare        InterventionType = "spiritual_care"
	TypeGeneticCounselor     InterventionType = "genetic_counselor"
	TypeRehabilitation       InterventionType = "rehabilitation"
	TypeClinicalTrial        InterventionType = "clinical_trial"
	TypeTranslationServices  InterventionType = "translation_services"
	TypePalliativeCare       InterventionType = "palliative_care"
	TypeHospice              InterventionType = "hospice"
	TypeSurvivorship         InterventionType = "survivorship"
	TypeCommunityResource    InterventionType = "community_resource"
	TypeCoordinationOfCare   InterventionType = "coordination_of_care"
	TypePhysicianOfficeStaff InterventionType = "physician_office_staff"
	TypeOther                InterventionType = "other"
)

type InterventionStatus string

const (
	StatusPending    InterventionStatus = "pending"
	StatusInProgress InterventionStatus = "in_progress"
	StatusCompleted  InterventionStatus = "completed"
	StatusCancelled  InterventionStatus = "cancelled"
)

type Intervention struct {
	ID              string             `gorm:"primaryKey;type:text" json:"id"`
	TenantID        string             `gorm:"type:text;index" json:"tenant_id"`
	PatientID       string             `gorm:"type:text;index" json:"patient_id"`
	ScreeningID     string             `gorm:"type:text;not null" json:"screening_id"`
	Type            InterventionType   `gorm:"type:text;not null" json:"type"`
	Title           string             `gorm:"type:text;not null" json:"title"`
	Description     *string            `gorm:"type:text" json:"description,omitempty"`
	Status          InterventionStatus `gorm:"type:text;not null;default:'pending'" json:"status"`
	Priority        string             `gorm:"type:text;default:'medium'" json:"priority"`
	CreatedBy       string             `gorm:"type:text;not null" json:"created_by"`
	AssignedTo      *string            `gorm:"type:text" json:"assigned_to,omitempty"`
	AssignedTeam    *string            `gorm:"type:text" json:"assigned_team,omitempty"`
	DueAt           *time.Time         `gorm:"type:timestamptz" json:"due_at,omitempty"`
	CompletedAt     *time.Time         `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	LinkedTaskID    *string            `gorm:"type:text" json:"linked_task_id,omitempty"`
	ReferralReasons pq.StringArray     `gorm:"type:text[]" json:"referral_reasons"`
	Problems        pq.StringArray     `gorm:"type:text[]" json:"problems"`
	Notes           *string            `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time          `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time          `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	User            *User              `gorm:"foreignKey:AssignedTo;references:ID" json:"user,omitempty"`
}

type BarrierCount struct {
	Month        string  `json:"month"`
	ProblemName  string  `json:"problem_name"`
	BarrierCount float64 `json:"barrier_count"`
}

type BarrierSubtype struct {
	SubType      string `json:"sub_type"`
	BarrierCount int    `json:"barrier_count"`
}

type BarrierResponse struct {
	ChartData   []*BarrierCount   `json:"chartData"`
	SubtypeData []*BarrierSubtype `json:"subtypeData"`
}

func (i *Intervention) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == "" {
		i.ID = "int_" + uuid.New().String()
	}
	if i.Status == "" {
		i.Status = StatusPending
	}
	return
}

func (i *Intervention) TableName() string {
	return "interventions"
}
