package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
)

type CognitoEvent struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	ChallengeAnswer      string            `json:"challengeAnswer"`
	PrivateChallengeParameters map[string]string `json:"privateChallengeParameters"`
}

type Response struct {
	AnswerCorrect bool `json:"answerCorrect"`
}

func HandleRequest(ctx context.Context, event CognitoEvent) (CognitoEvent, error) {
	expectedAnswer, ok := event.Request.PrivateChallengeParameters["otp"]
	if !ok {
		return event, errors.New("OTP not found in challenge parameters")
	}

	if event.Request.ChallengeAnswer == expectedAnswer {
		event.Response.AnswerCorrect = true
	} else {
		event.Response.AnswerCorrect = false
	}

	return event, nil
}

func main() {
	lambda.Start(HandleRequest)
}
