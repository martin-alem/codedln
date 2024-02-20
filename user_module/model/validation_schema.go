package model

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"codedln/util/helpers"
	"codedln/util/types"
	"errors"
	"net/http"
)

type CreateUserSchema struct {
	IdToken    string
	SignInWith types.OAuthSignIn
}

func NewCreateUserSchema() CreateUserSchema {
	return CreateUserSchema{}
}

func (s CreateUserSchema) Validate() error {
	if s.IdToken == "" {
		return errors.New("id token must exist")
	}
	oAuthList := []types.OAuthSignIn{constant.GoogleSignIn, constant.GitHubSignIn}
	if !helpers.InList(oAuthList, s.SignInWith, func(a types.OAuthSignIn, b types.OAuthSignIn) bool {
		return a == b
	}) {
		return http_error.New(http.StatusBadRequest, "invalid sign in method")
	}

	return nil
}
