package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestHandler_SetMetricsUniversal(t *testing.T) {
	t.Parallel()

	const defaultMetricName = "someMetricName"

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		wantStatus  int
		setupMock   func(service *mocks.MockMetricService)
	}{
		{
			name:        "gauge ok",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge","value":1.124}`,
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					SetGauge(gomock.Any(), defaultMetricName, 1.124).
					Return(nil)
			},
		},
		{
			name:        "counter ok",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"counter","delta":10}`,
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					IncrementCounter(gomock.Any(), defaultMetricName, int64(10)).
					Return(nil)
			},
		},
		{
			name:        "method not allowed",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge","value":1.124}`,
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name:        "unsupported content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        `{"id":"someMetricName","type":"gauge","value":1.124}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid json",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{invalid json}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid metric type",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"histogram","value":1.124}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric id",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"","type":"gauge","value":1.124}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "nil gauge value",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "nil counter delta",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"counter"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "gauge service error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge","value":1.124}`,
			wantStatus:  http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					SetGauge(gomock.Any(), defaultMetricName, 1.124).
					Return(errors.New("set gauge failed"))
			},
		},
		{
			name:        "counter service error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"counter","delta":10}`,
			wantStatus:  http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					IncrementCounter(gomock.Any(), defaultMetricName, int64(10)).
					Return(errors.New("increment counter failed"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "http://sample.ru/update/", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockMetricService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			logger := zaptest.NewLogger(t)
			h := NewHandler(service, logger)

			h.SetMetricsUniversal(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
