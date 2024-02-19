package helpers

import (
	"codedln/shared/http_error"
	"encoding/json"
	"net"
	"net/http"
)

func GetClientIP(req *http.Request) string {
	// Standard headers used by Amazon ELB, Heroku, and others.
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback to using the remote address from the request.
	// This will give the network IP, which might be a proxy.
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

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
