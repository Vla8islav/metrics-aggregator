package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestHandler_GetMetrics(t *testing.T) {
	t.Parallel()

	const defaultMetricName = "someMetricName"

	tests := []struct {
		name       string
		method     string
		vars       map[string]string
		wantStatus int
		wantBody   string
		setupMock  func(service *mocks.MockMetricService)
	}{
		{
			name:   "gauge ok",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Gauge),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusOK,
			wantBody:   "1.124",
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(1.124, nil)
			},
		},
		{
			name:   "counter ok",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Counter),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusOK,
			wantBody:   "10",
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetCounter(gomock.Any(), defaultMetricName).
					Return(int64(10), nil)
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			vars: map[string]string{
				"metricType": string(models.Gauge),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid metric type",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": "histogram",
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "empty metric name",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Gauge),
				"metricName": "",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "gauge not found",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Gauge),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusNotFound,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(0.0, repository.ErrNotFound)
			},
		},
		{
			name:   "gauge service error",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Gauge),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(0.0, errors.New("get gauge failed"))
			},
		},
		{
			name:   "counter not found",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Counter),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusNotFound,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetCounter(gomock.Any(), defaultMetricName).
					Return(int64(0), repository.ErrNotFound)
			},
		},
		{
			name:   "counter service error",
			method: http.MethodGet,
			vars: map[string]string{
				"metricType": string(models.Counter),
				"metricName": defaultMetricName,
			},
			wantStatus: http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetCounter(gomock.Any(), defaultMetricName).
					Return(int64(0), errors.New("get counter failed"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := newReqWithVars(t, "http://sample.ru/value/", tt.method, tt.vars, "")
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockMetricService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			logger := zaptest.NewLogger(t)
			h := NewHandler(service, logger)

			h.GetMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantBody != "" {
				require.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
