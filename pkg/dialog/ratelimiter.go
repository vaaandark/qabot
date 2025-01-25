package dialog

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Second), 20)

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests, please try again later", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
