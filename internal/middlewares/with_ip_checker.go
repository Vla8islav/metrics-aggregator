package middlewares

import (
	"net/http"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"go.uber.org/zap"
)

// WithIpChecker checks if the request came from the correct subnet
//
// example: trusted_subnet is 192.168.88.0/24 and X-Real-IP in request is 192.168.88.15 - PASS
func WithIpChecker(subnet config.OptionalString, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !subnet.BeenSet {
				next.ServeHTTP(w, r)
				return
			}

			requestIP := r.Header.Get(helpers.RealIpHeader)

			if requestIP == "" {
				logger.Error("missing X-Real-IP header")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			if err2 := helpers.ValidateIPByCIDR(requestIP, subnet.Value); err2 != nil {
				logger.Error("invalid X-Real-IP header server doesn't accept requests from this subnet", zap.Error(err2))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)

		})

	}
}
