package middlewares

import (
	"net/http"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/audit"
)
import _ "net/http/pprof"

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
