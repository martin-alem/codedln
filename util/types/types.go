package types

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type HTTPHandler func(http.ResponseWriter, *http.Request) error

type Middleware func(HTTPHandler) HTTPHandler

type ValidatableSchema interface {
	Validate() error
}

type PayloadKey struct{}

type OAuthSignIn string

type ContextKey int

type JWTClaim struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

type AuthUser struct {
	UserId string `json:"userId"`
}

type ServerResponse struct {
	Message    string `json:"message"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"statusCode"`
	TimeStamp  string `json:"timestamp"`
}

type GoogleOAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IdToken     string `json:"id_token"`
}

type PaginationResult[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
}
