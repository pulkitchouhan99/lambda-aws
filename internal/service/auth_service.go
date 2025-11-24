package service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type AuthService struct {
	cognitoClient *cognitoidentityprovider.Client
	userPoolID    string
	clientID      string
}

func NewAuthService(cfg aws.Config, userPoolID, clientID string) *AuthService {
	client := cognitoidentityprovider.NewFromConfig(cfg)
	return &AuthService{
		cognitoClient: client,
		userPoolID:    userPoolID,
		clientID:      clientID,
	}
}

func (s *AuthService) CreateCognitoUser(email, password, role, tenantID, navigatorAdminID string) (*string, error) {
	resp, err := s.cognitoClient.SignUp(context.TODO(), &cognitoidentityprovider.SignUpInput{
		ClientId: &s.clientID,
		Password: &password,
		Username: &email,
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("custom:role"), Value: aws.String(role)},
			{Name: aws.String("custom:tenant_id"), Value: aws.String(tenantID)},
			{Name: aws.String("custom:navigator_admin_id"), Value: aws.String(navigatorAdminID)},
		},
	})
	if err != nil {
		return nil, err
	}
	return resp.UserSub, nil
}

func (s *AuthService) StartOTPChallenge(email, password string) (*cognitoidentityprovider.RespondToAuthChallengeOutput, error) {
	resp, err := s.cognitoClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeCustomAuth,
		ClientId: &s.clientID,
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return nil, err
	}

	challengeResp, err := s.cognitoClient.RespondToAuthChallenge(context.TODO(), &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName: types.ChallengeNameTypeCustomChallenge,
		ClientId:      &s.clientID,
		ChallengeResponses: map[string]string{
			"USERNAME": email,
		},
		Session: resp.Session,
	})
	if err != nil {
		return nil, err
	}

	return challengeResp, nil
}

func (s *AuthService) VerifyOTPChallenge(email, otp, session string) (*types.AuthenticationResultType, error) {
	resp, err := s.cognitoClient.RespondToAuthChallenge(context.TODO(), &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName: types.ChallengeNameTypeCustomChallenge,
		ClientId:      &s.clientID,
		ChallengeResponses: map[string]string{
			"USERNAME": email,
			"ANSWER":   otp,
		},
		Session: &session,
	})
	if err != nil {
		return nil, err
	}

	return resp.AuthenticationResult, nil
}
