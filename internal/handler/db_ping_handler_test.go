package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/metrics-aggregator/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestHandler_DBPing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantStatus int
		serviceErr error
	}{
		{
			name:       "get ok",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "service ping error",
			method:     http.MethodGet,
			wantStatus: http.StatusInternalServerError,
			serviceErr: errors.New("ping failed"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "http://sample.ru/ping", nil)
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := mocks.NewMockMetricService(ctrl)
			if tt.method == http.MethodGet {
				service.EXPECT().
					Ping(gomock.Any()).
					Return(tt.serviceErr)
			}

			logger := zaptest.NewLogger(t)
			h := NewHandler(service, logger)

			h.DBPing(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
