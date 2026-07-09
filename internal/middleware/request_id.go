package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func RequestID(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := fmt.Sprintf("%d", time.Now().UnixNano())
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}
