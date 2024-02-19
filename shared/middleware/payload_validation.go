package middleware

import (
	"codedln/shared/http_error"
	"codedln/util/constant"
	"codedln/util/types"
	"context"
	"encoding/json"
	"net/http"
)

func PayloadValidationMiddleware[T types.ValidatableSchema](factory func() T) types.Middleware {
	return func(handler types.HTTPHandler) types.HTTPHandler {
		return func(w http.ResponseWriter, r *http.Request) error {

			r.Body = http.MaxBytesReader(w, r.Body, constant.MaxPayloadSize)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields() // return an error if extra fields are present

			payloadSchema := factory()

			if err := decoder.Decode(&payloadSchema); err != nil {
				return http_error.New(400, "invalid payload")
			}

			// Validate fields
			if err := payloadSchema.Validate(); err != nil {
				return http_error.New(400, "Validation error: "+err.Error())
			}

			ctx := context.WithValue(r.Context(), types.PayloadKey{}, payloadSchema)
			return handler(w, r.WithContext(ctx))
		}
	}
}
