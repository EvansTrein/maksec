package rest

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func LogMiddleware(log *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			l := log.With().Str("request_id", uuid.New().String()).Logger()
			defer func() {
				stop := time.Now()

				l.Info().
					Str("remote_ip", r.RemoteAddr).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("user_agent", r.UserAgent()).
					Dur("latency", stop.Sub(start)).
					Msg("request processed")
			}()

			next.ServeHTTP(w, r)
		})
	}
}
