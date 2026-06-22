package handler

import (
	"bytes"
	"encoding/json"
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

func TestHandler_GetMetricsUniversal(t *testing.T) {
	t.Parallel()

	const defaultMetricName = "someMetricName"

	tests := []struct {
		name            string
		method          string
		contentType     string
		body            string
		wantStatus      int
		wantContentType string
		wantResponse    models.Metrics
		setupMock       func(service *mocks.MockMetricService)
	}{
		{
			name:            "gauge ok",
			method:          http.MethodPost,
			contentType:     "application/json",
			body:            `{"id":"someMetricName","type":"gauge"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponse: models.Metrics{
				ID:    defaultMetricName,
				MType: models.Gauge,
				Value: func() *float64 {
					v := 1.124
					return &v
				}(),
			},
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(1.124, nil)
			},
		},
		{
			name:            "counter ok",
			method:          http.MethodPost,
			contentType:     "application/json",
			body:            `{"id":"someMetricName","type":"counter"}`,
			wantStatus:      http.StatusOK,
			wantContentType: "application/json",
			wantResponse: models.Metrics{
				ID:    defaultMetricName,
				MType: models.Counter,
				Delta: func() *int64 {
					v := int64(10)
					return &v
				}(),
			},
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetCounter(gomock.Any(), defaultMetricName).
					Return(int64(10), nil)
			},
		},
		{
			name:        "method not allowed",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge"}`,
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name:        "unsupported content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        `{"id":"someMetricName","type":"gauge"}`,
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
			body:        `{"id":"someMetricName","type":"histogram"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "empty metric id",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"","type":"gauge"}`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "gauge not found",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge"}`,
			wantStatus:  http.StatusNotFound,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(0.0, repository.ErrNotFound)
			},
		},
		{
			name:        "gauge service error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"gauge"}`,
			wantStatus:  http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetGauge(gomock.Any(), defaultMetricName).
					Return(0.0, errors.New("get gauge failed"))
			},
		},
		{
			name:        "counter not found",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"counter"}`,
			wantStatus:  http.StatusNotFound,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetCounter(gomock.Any(), defaultMetricName).
					Return(int64(0), repository.ErrNotFound)
			},
		},
		{
			name:        "counter service error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"id":"someMetricName","type":"counter"}`,
			wantStatus:  http.StatusInternalServerError,
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

			req := httptest.NewRequest(tt.method, "http://sample.ru/value/", bytes.NewBufferString(tt.body))
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

			h.GetMetricsUniversal(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, rr.Header().Get("Content-Type"))
			}

			if tt.wantResponse.ID != "" {
				var got models.Metrics
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err)
				require.Equal(t, tt.wantResponse.ID, got.ID)
				require.Equal(t, tt.wantResponse.MType, got.MType)

				if tt.wantResponse.Value != nil {
					require.NotNil(t, got.Value)
					require.InDelta(t, *tt.wantResponse.Value, *got.Value, 1e-9)
				}

				if tt.wantResponse.Delta != nil {
					require.NotNil(t, got.Delta)
					require.Equal(t, *tt.wantResponse.Delta, *got.Delta)
				}
			}
		})
	}
}
