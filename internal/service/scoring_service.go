package service

import "github.com/lambda/internal/domain"

type ScoringService struct{}

func NewScoringService() *ScoringService {
	return &ScoringService{}
}

func (s *ScoringService) CalculateNCCNScore(screening *domain.Screening) int {
	// This is a stub. In a real application, this would contain
	// the logic to calculate the NCCN score based on the screening answers.
	return 100
}
