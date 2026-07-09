package middleware

import (
	"log/slog"
	"net/http"
)

func Recovery(next http.Handler, logger *slog.Logger) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				logger.Error(
					"Panic recovered",
					"error", err,
				)

				http.Error(
					w,
					"Internal Server Error",
					http.StatusInternalServerError,
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
