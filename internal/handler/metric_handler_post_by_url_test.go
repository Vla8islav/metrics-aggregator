package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func newReqWithVars(t *testing.T, target string, method string, vars map[string]string, contentType string) *http.Request {
	t.Helper()
	//
	req := httptest.NewRequest(method, target, nil)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return mux.SetURLVars(req, vars)
}

func TestPostMetrics(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	const defaultMetricName = "someMetricName"

	tests := []struct {
		name        string
		vars        map[string]string
		contentType string
		wantStatus  int
	}{
		{
			name: "gauge ok", vars: map[string]string{
				"metricType":  string(models.Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "1.124",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
		},
		{
			name: "counter ok", vars: map[string]string{
				"metricType":  string(models.Counter),
				"metricName":  defaultMetricName,
				"metricValue": "1",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
		},
		{
			name: "invalid metric type",
			vars: map[string]string{
				"metricType":  "histogram",
				"metricName":  "any",
				"metricValue": "1",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "empty metric name",
			vars: map[string]string{
				"metricType":  string(models.Gauge),
				"metricName":  "",
				"metricValue": "1.23",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "empty metric value",
			vars: map[string]string{
				"metricType":  string(models.Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "gauge parse error",
			vars: map[string]string{
				"metricType":  string(models.Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "invalid_value",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "counter parse error",
			vars: map[string]string{
				"metricType":  string(models.Counter),
				"metricName":  defaultMetricName,
				"metricValue": "1.2",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			req := newReqWithVars(t, "http://sample.ru/update/", http.MethodPost, tt.vars, tt.contentType)
			rr := httptest.NewRecorder()

			metricName := tt.vars["metricName"]

			cfg := config.ReadFlags(nil)
			db := repository.NewMemStorage(cfg)
			zap := zaptest.NewLogger(t)
			h := NewHandler(db, zap)
			h.SetMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				expectedMetricValueStr := tt.vars["metricValue"]
				switch tt.vars["metricType"] {
				case string(models.Gauge):
					metricValue, err := strconv.ParseFloat(expectedMetricValueStr, 64)
					require.NoError(t, err)
					val, err := h.service.GetGauge(ctx, metricName)
					require.NoError(t, err)
					assert.InDelta(t, metricValue, val, 1e-9)
				case string(models.Counter):
					metricValue, err := strconv.ParseInt(expectedMetricValueStr, 10, 64)
					require.NoError(t, err)
					val, err := h.service.GetCounter(ctx, metricName)
					require.NoError(t, err)
					assert.EqualValues(t, metricValue, val, 1e-9)
				}
			}
		})

	}

}
