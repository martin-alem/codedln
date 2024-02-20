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
