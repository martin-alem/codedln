package middleware

import (
	"codedln/shared/http_error"
	"codedln/util/helpers"
	"codedln/util/types"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"net/http"
	"strconv"
	"time"
)

func RateLimitMiddleware(rateLimiter *redis_rate.Limiter, limit redis_rate.Limit) types.Middleware {
	return func(next types.HTTPHandler) types.HTTPHandler {
		return func(w http.ResponseWriter, r *http.Request) error {

			clientIP := helpers.GetClientIP(r)
			key := fmt.Sprintf("%s:%s:%s:%s:%s", "rate_limit", clientIP, r.RequestURI, r.Host, r.Method)

			res, err := rateLimiter.Allow(r.Context(), key, limit)
			if err != nil {
				return err
			}

			h := w.Header()
			h.Set("RateLimit-Remaining", strconv.Itoa(res.Remaining))

			if res.Allowed == 0 {
				seconds := int(res.RetryAfter / time.Second)
				h.Set("RateLimit-RetryAfter", strconv.Itoa(seconds))
				return http_error.New(429, "too many request")
			}
			return next(w, r)
		}
	}
}
