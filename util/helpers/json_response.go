package helpers

import (
	"codedln/shared/http_error"
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response with a given status code.
func JSONResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			return http_error.New(500, "internal server error")
		}
	}
	return nil
}

// JSONDecode decodes a JSON request into a struct.
func JSONDecode(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dst)
}
