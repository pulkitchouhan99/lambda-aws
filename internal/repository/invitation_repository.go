package repository

import (
	"github.com/lambda/internal/domain"
	"gorm.io/gorm"
)

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) CreateInvitation(invitation *domain.Invitation) error {
	return r.db.Create(invitation).Error
}

func (r *InvitationRepository) FindInvitationByToken(token string) (*domain.Invitation, error) {
	var invitation domain.Invitation
	if err := r.db.Where("token = ? AND is_used = ?", token, false).First(&invitation).Error; err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *InvitationRepository) UpdateInvitationStatus(token string, isUsed bool) error {
	return r.db.Model(&domain.Invitation{}).Where("token = ?", token).Update("is_used", isUsed).Error
}
