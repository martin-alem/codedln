package middleware

import (
	"codedln/util/types"
	"net/http"
)

func CorsMiddleware(next types.HTTPHandler) types.HTTPHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Check if it is a preflight request
		if r.Method == http.MethodOptions {
			// Respond to the preflight request with the necessary headers
			w.WriteHeader(http.StatusOK)
			return nil
		}

		// Call the next handler in the chain
		return next(w, r)
	}
}
