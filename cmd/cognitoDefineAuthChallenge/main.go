package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

type CognitoEvent struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Session []ChallengeResult `json:"session"`
}

type ChallengeResult struct {
	ChallengeName   string `json:"challengeName"`
	ChallengeResult bool   `json:"challengeResult"`
}

type Response struct {
	ChallengeName string `json:"challengeName"`
	Fail          bool   `json:"failAuthentication"`
	IssueTokens   bool   `json:"issueTokens"`
}

func HandleRequest(ctx context.Context, event CognitoEvent) (CognitoEvent, error) {
	if len(event.Request.Session) == 0 {
		// Start with password verification
		event.Response.ChallengeName = "CUSTOM_CHALLENGE"
		event.Response.Fail = false
		event.Response.IssueTokens = false
	} else {
		// After password verification, move to OTP
		lastChallenge := event.Request.Session[len(event.Request.Session)-1]
		if lastChallenge.ChallengeName == "CUSTOM_CHALLENGE" && lastChallenge.ChallengeResult == true {
			event.Response.ChallengeName = "CUSTOM_CHALLENGE"
			event.Response.Fail = false
			event.Response.IssueTokens = false
		} else {
			event.Response.Fail = true
			event.Response.IssueTokens = false
		}
	}

	// If OTP is verified, issue tokens
	if len(event.Request.Session) > 1 {
		otpChallenge := event.Request.Session[len(event.Request.Session)-1]
		if otpChallenge.ChallengeName == "CUSTOM_CHALLENGE" && otpChallenge.ChallengeResult == true {
			event.Response.IssueTokens = true
			event.Response.Fail = false
		}
	}

	return event, nil
}

func main() {
	lambda.Start(HandleRequest)
}
