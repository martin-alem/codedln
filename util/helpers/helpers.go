package helpers

import (
	"codedln/shared/http_error"
	"codedln/util/types"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
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

func InList[T any](list []T, target T, predicate func(a T, b T) bool) bool {
	for _, v := range list {
		if predicate(v, target) == true {
			return true
		}
	}
	return false
}

func ternary[T any](expr bool, result1 T, result2 T) T {
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

func CreateCook(name string, value string, ttl int) http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(ttl) * time.Hour),
		HttpOnly: ternary[bool](os.Getenv("ENVIRONMENT") == "production", true, false),
		SameSite: http.SameSiteStrictMode,
		Secure:   ternary[bool](os.Getenv("ENVIRONMENT") == "production", true, false),
	}

	return cookie
}
