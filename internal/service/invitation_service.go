package service

import (
	"github.com/lambda/internal/domain"
	"github.com/lambda/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type InvitationService struct {
	repo      *repository.InvitationRepository
	jwtSecret []byte
}

func NewInvitationService(repo *repository.InvitationRepository, jwtSecret string) *InvitationService {
	return &InvitationService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *InvitationService) CreateInvitation(invitation *domain.Invitation) (string, error) {
	claims := jwt.MapClaims{
		"email": invitation.Email,
		"role":  invitation.Role,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	invitation.Token = tokenString
	invitation.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)
	if err := s.repo.CreateInvitation(invitation); err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *InvitationService) ValidateInvitationToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
