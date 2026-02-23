package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReqWithVars(t *testing.T, method string, vars map[string]string, contentType string) *http.Request {
	t.Helper()
	//
	req := httptest.NewRequest(method, "http://sample.ru/update/", nil)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return mux.SetURLVars(req, vars)
}

func TestPostMetrics(t *testing.T) {
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
			"metricType":  string(Gauge),
			"metricName":  defaultMetricName,
			"metricValue": "1.124",
		},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
		},
		{
			name: "counter ok", vars: map[string]string{
			"metricType":  string(Counter),
			"metricName":  defaultMetricName,
			"metricValue": "1",
		},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
		},
		{
			name: "bad content type", vars: map[string]string{
			"metricType":  string(Gauge),
			"metricName":  defaultMetricName,
			"metricValue": "1.488",
		},
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
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
				"metricType":  string(Gauge),
				"metricName":  "",
				"metricValue": "1.23",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "empty metric value",
			vars: map[string]string{
				"metricType":  string(Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "gauge parse error",
			vars: map[string]string{
				"metricType":  string(Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "invalid_value",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "counter parse error",
			vars: map[string]string{
				"metricType":  string(Counter),
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

			req := newReqWithVars(t, http.MethodPost, tt.vars, tt.contentType)
			rr := httptest.NewRecorder()

			metricName := tt.vars["metricName"]

			PostMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantStatus == http.StatusOK {
				expectedMetricValueStr := tt.vars["metricValue"]
				switch tt.vars["metricType"] {
				case string(Gauge):
					metricValue, err := strconv.ParseFloat(expectedMetricValueStr, 64)
					require.NoError(t, err)
					val, err := repository.MemStorage.GetGauge(metricName)
					require.NoError(t, err)
					assert.InDelta(t, metricValue, val, 1e-9)
				case string(Counter):
					metricValue, err := strconv.ParseInt(expectedMetricValueStr, 10, 64)
					require.NoError(t, err)
					val, err := repository.MemStorage.GetCounter(metricName)
					require.NoError(t, err)
					assert.EqualValues(t, metricValue, val, 1e-9)
				}
			}
		})

	}

}
