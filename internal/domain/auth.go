package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
		RolePatient            Role = "patient"
		RolePatientNavigator   Role = "patient_navigator"
		RoleSocialWorker       Role = "social_worker"
		RoleNavigatorAdmin     Role = "navigator_admin"
		RoleNurseNavigator     Role = "nurse_navigator"
	RoleRegisteredDietitian Role = "registered_dietitian"
)

type User struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	TenantID         uuid.UUID `json:"tenant_id" gorm:"type:uuid"`
	Email            string    `json:"email" gorm:"uniqueIndex"`
	Username         string    `json:"username" gorm:"uniqueIndex"`
	PhoneNumber      string    `json:"phone_number"`
	Role             Role      `json:"role"`
	NavigatorAdminID uuid.UUID `json:"navigator_admin_id" gorm:"type:uuid"`
	IsDeleted        bool      `json:"is_deleted"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Invitation struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	Role      Role      `json:"role"`
	InvitedBy uuid.UUID `json:"invited_by" gorm:"type:uuid"`
	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}
