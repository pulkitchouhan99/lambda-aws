package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"math/big"
)

type CognitoEvent struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	UserAttributes map[string]string `json:"userAttributes"`
	Session        []ChallengeResult `json:"session"`
}

type ChallengeResult struct {
	ChallengeName    string `json:"challengeName"`
	ChallengeResult  bool   `json:"challengeResult"`
	ChallengeMetadata *string `json:"challengeMetadata"`
}

type Response struct {
	PublicChallengeParameters  map[string]string `json:"publicChallengeParameters"`
	PrivateChallengeParameters map[string]string `json:"privateChallengeParameters"`
	ChallengeMetadata          string            `json:"challengeMetadata"`
}

func HandleRequest(ctx context.Context, event CognitoEvent) (CognitoEvent, error) {
	otp, err := generateOTP()
	if err != nil {
		return event, err
	}

	event.Response.PrivateChallengeParameters = map[string]string{"otp": otp}
	event.Response.PublicChallengeParameters = map[string]string{"email": event.Request.UserAttributes["email"]}
	event.Response.ChallengeMetadata = "OTP_CHALLENGE"

	// In a real application, you would send the OTP via email or SMS here

	return event, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func main() {
	lambda.Start(HandleRequest)
}
