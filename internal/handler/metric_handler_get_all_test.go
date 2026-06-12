package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestHandler_GetAllMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		method          string
		wantStatus      int
		wantContentType string
		wantBodyParts   []string
		setupMock       func(service *mocks.MockMetricService)
	}{
		{
			name:            "get all metrics ok",
			method:          http.MethodGet,
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
			wantBodyParts: []string{
				"<h1>Metrics</h1>",
				"<h2>Gauges</h2>",
				"<strong>Alloc</strong>: 1.124",
				"<strong>RandomValue</strong>: 42",
				"<h2>Counters</h2>",
				"<strong>PollCount</strong>: 10",
			},
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetAll(gomock.Any()).
					Return(models.MetricsExport{
						Gauges: map[string]float64{
							"Alloc":       1.124,
							"RandomValue": 42,
						},
						Counters: map[string]int64{
							"PollCount": 10,
						},
					}, nil)
			},
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "service get all error",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetAll(gomock.Any()).
					Return(models.MetricsExport{}, errors.New("get all failed"))
			},
		},
		{
			name:            "empty metrics ok",
			method:          http.MethodGet,
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
			wantBodyParts: []string{
				"<h1>Metrics</h1>",
				"No gauges available.",
				"No counters available.",
			},
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					GetAll(gomock.Any()).
					Return(models.MetricsExport{}, nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "http://sample.ru/", nil)
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockMetricService(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			logger := zaptest.NewLogger(t)
			h := NewHandler(service, logger)

			h.GetAllMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, rr.Header().Get("Content-Type"))
			}

			for _, part := range tt.wantBodyParts {
				require.Contains(t, rr.Body.String(), part)
			}
		})
	}
}
