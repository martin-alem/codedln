package middleware

import (
	"codedln/shared/http_error"
	"codedln/util/types"
	"net/http"
	"os"
	"strings"
)

func ValidateAPIKeyMiddleware(next types.HTTPHandler) types.HTTPHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		header := r.Header.Get("Authorization")
		if header == "" {
			return http_error.New(http.StatusUnauthorized, "authorization api key required")
		}

		headerSlice := strings.Split(header, " ")

		if len(headerSlice) < 2 {
			return http_error.New(http.StatusBadRequest, "malformed client authorization header")
		}

		if strings.Compare(headerSlice[1], os.Getenv("CLIENT_KEY")) != 0 {
			return http_error.New(http.StatusUnauthorized, "invalid client authorization token")
		}

		return next(w, r)
	}
}
