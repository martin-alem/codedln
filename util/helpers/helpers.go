package helpers

import (
	"bytes"
	"codedln/shared/http_error"
	"codedln/util/types"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net"
	"net/http"
	"os"
	"time"
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
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.WriteHeader(statusCode)

	var payload any
	if statusCode >= 400 {
		payload = data
	} else {
		payload = map[string]any{
			"status_code": statusCode,
			"data":        data,
		}
	}

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return http_error.New(500, "internal server error")
	}

	return nil
}

// JSONDecode decodes a JSON request into a struct.
func JSONDecode(r io.Reader, dst any) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(dst)
}

func InList[T any](list []T, target T, predicate func(a T, b T) bool) bool {
	for _, v := range list {
		if predicate(v, target) == true {
			return true
		}
	}
	return false
}

func Ternary[T any](expr bool, result1 T, result2 T) T {
	if expr == true {
		return result1
	} else {
		return result2
	}
}

func CreateJWT(claim types.AuthUser, ttl int) string {
	claims := types.JWTClaim{
		UserId: claim.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Codedln",
			Subject:   "Access Token",
			ID:        "1",
			Audience:  []string{"Codedln"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return ""
	}

	return ss
}

func CreateCookie(name string, value string, ttl int) http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(ttl) * time.Hour),
		HttpOnly: Ternary[bool](os.Getenv("ENVIRONMENT") == "production", true, false),
		SameSite: http.SameSiteStrictMode,
		Secure:   Ternary[bool](os.Getenv("ENVIRONMENT") == "production", true, false),
	}

	return cookie
}

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		data := map[string]interface{}{"message": "Resource Not Found", "statusCode": http.StatusNotFound, "path": r.URL.Path, "method": r.Method, "timestamp": time.Now().UTC()}
		_ = json.NewEncoder(w).Encode(data)
	}
}

func MethodNotAllowed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		data := map[string]interface{}{"message": "Method Not Allowed", "statusCode": http.StatusMethodNotAllowed, "path": r.URL.Path, "method": r.Method, "timestamp": time.Now().UTC()}
		_ = json.NewEncoder(w).Encode(data)
	}
}

func PreflightRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
		return
	}
}

func AnyTypeToReader(data interface{}) (io.Reader, error) {
	// Serialize the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err // Handle serialization errors
	}

	// Create a bytes.Reader from the serialized data
	reader := bytes.NewReader(jsonData)

	return reader, nil
}
