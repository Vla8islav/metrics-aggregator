package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func newReqWithVars(t *testing.T, target string, method string, vars map[string]string, contentType string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
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
		setupMock   func(service *mocks.MockMetricService)
	}{
		{
			name: "gauge ok", vars: map[string]string{
				"metricType":  string(models.Gauge),
				"metricName":  defaultMetricName,
				"metricValue": "1.124",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					SetGauge(gomock.Any(), defaultMetricName, 1.124).
					Return(nil)
			},
		},
		{
			name: "counter ok", vars: map[string]string{
				"metricType":  string(models.Counter),
				"metricName":  defaultMetricName,
				"metricValue": "1",
			},
			contentType: "text/plain",
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					IncrementCounter(gomock.Any(), defaultMetricName, int64(1)).
					Return(nil)
			},
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
			t.Parallel()

			req := newReqWithVars(t, "http://sample.ru/update/", http.MethodPost, tt.vars, tt.contentType)
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockMetricService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			logger := zaptest.NewLogger(t)
			h := NewHandler(service, logger)

			h.SetMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
