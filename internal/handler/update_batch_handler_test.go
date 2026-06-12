package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestHandler_UpdateBatchMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		wantStatus  int
		setupMock   func(service *mocks.MockMetricService)
	}{
		{
			name:        "batch update ok",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `[{"id":"Alloc","type":"gauge","value":1.124},{"id":"PollCount","type":"counter","delta":10}]`,
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				gaugeValue := 1.124
				counterDelta := int64(10)

				service.EXPECT().
					UpdateMetrics(gomock.Any(), []models.Metrics{
						{
							ID:    "Alloc",
							MType: models.Gauge,
							Value: &gaugeValue,
						},
						{
							ID:    "PollCount",
							MType: models.Counter,
							Delta: &counterDelta,
						},
					}).
					Return(nil)
			},
		},
		{
			name:        "empty batch ok",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `[]`,
			wantStatus:  http.StatusOK,
			setupMock: func(service *mocks.MockMetricService) {
				service.EXPECT().
					UpdateMetrics(gomock.Any(), []models.Metrics{}).
					Return(nil)
			},
		},
		{
			name:        "method not allowed",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        `[]`,
			wantStatus:  http.StatusMethodNotAllowed,
		},
		{
			name:        "unsupported content type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        `[]`,
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
			name:        "service update error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `[{"id":"Alloc","type":"gauge","value":1.124}]`,
			wantStatus:  http.StatusInternalServerError,
			setupMock: func(service *mocks.MockMetricService) {
				gaugeValue := 1.124

				service.EXPECT().
					UpdateMetrics(gomock.Any(), []models.Metrics{
						{
							ID:    "Alloc",
							MType: models.Gauge,
							Value: &gaugeValue,
						},
					}).
					Return(errors.New("update metrics failed"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "http://sample.ru/updates/", bytes.NewBufferString(tt.body))
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

			h.UpdateBatchMetrics(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
