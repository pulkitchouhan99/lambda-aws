package events

import (
	"github.com/google/uuid"
	"time"
)

type UserInvited struct {
	InvitationID uuid.UUID `json:"invitation_id"`
	Email        string    `json:"email"`
	InvitedBy    uuid.UUID `json:"invited_by"`
	InvitedAt    time.Time `json:"invited_at"`
}

type UserRegistered struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type UserOtpSent struct {
	UserID    uuid.UUID `json:"user_id"`
	Method    string    `json:"method"` // e.g., "email", "sms"
	Timestamp time.Time `json:"timestamp"`
}

type UserLoggedIn struct {
	UserID    uuid.UUID `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
