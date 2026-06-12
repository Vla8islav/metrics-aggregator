package middlewares

import (
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
)

// WithAudit returns middleware that records request audit data and publishes it after the handler runs.
//
// The middleware adds audit request data to the request context so handlers can fill in
// operation and metric details before the audit event is published.
func WithAudit(publisher *audit.Publisher) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			data := audit.NewRequestData()
			ctx := audit.WithRequestData(r.Context(), data)
			r = r.WithContext(ctx)

			start := time.Now()

			next.ServeHTTP(w, r)

			event := audit.Event{
				Operation:  data.Operation,
				Time:       start,
				Metrics:    data.Metrics,
				RemoteAddr: r.RemoteAddr,
			}

			err := publisher.Publish(r.Context(), event)
			if err != nil {
				return
			}
		})
	}
}
