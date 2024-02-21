package middleware

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"codedln/util/types"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

func AuthenticationMiddleware(next types.HTTPHandler) types.HTTPHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		header := r.Header.Get("Authorization")
		if header == "" {
			return http_error.New(http.StatusUnauthorized, "authentication required")
		}

		headerSlice := strings.Split(header, " ")

		if len(headerSlice) < 2 {
			return http_error.New(http.StatusBadRequest, "malformed client authentication header")
		}

		if strings.Compare(headerSlice[1], os.Getenv("CLIENT_KEY")) != 0 {
			return http_error.New(http.StatusUnauthorized, "invalid client authentication token")
		}

		cookie, cookieErr := r.Cookie(constant.JwtCookieName)
		switch {
		case errors.Is(cookieErr, http.ErrNoCookie):
			return http_error.New(http.StatusUnauthorized, "no authentication cookie provided")
		case cookieErr != nil:
			return http_error.New(http.StatusUnauthorized, "must be authenticated")
		}

		err := cookie.Valid()
		if err != nil {
			return http_error.New(http.StatusUnauthorized, "cookie expired")
		}

		accessToken := cookie.Value
		jwtClaim := &types.JWTClaim{}
		token, tokenErr := jwt.ParseWithClaims(accessToken, jwtClaim, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		switch {
		case errors.Is(tokenErr, jwt.ErrTokenExpired):
			return http_error.New(http.StatusUnauthorized, "access token expired")
		case tokenErr != nil:
			return http_error.New(http.StatusUnauthorized, "invalid authentication token")
		}

		var ctx context.Context
		if claim, ok := token.Claims.(*types.JWTClaim); ok {
			ctx = context.WithValue(r.Context(), constant.AuthUserKey, types.AuthUser{UserId: claim.UserId})
		}

		return next(w, r.WithContext(ctx))
	}
}
