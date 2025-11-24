package graph

import (
	"github.com/lambda/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	AuthService       *service.AuthService
	InvitationService *service.InvitationService
}