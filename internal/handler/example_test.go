package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"go.uber.org/zap"
)

type fakeService struct{}

func (fakeService) GetAll(ctx context.Context) (models.MetricsExport, error) {
	return models.MetricsExport{
		Gauges: map[string]float64{
			"Alloc": 123.45,
		},
		Counters: map[string]int64{
			"PollCount": 10,
		},
	}, nil
}

func (fakeService) Ping(ctx context.Context) error {
	return nil
}

func (fakeService) IncrementCounter(ctx context.Context, name string, number int64) error {
	return nil
}

func (fakeService) SetGauge(ctx context.Context, name string, gauge float64) error {
	return nil
}

func (fakeService) GetGauge(ctx context.Context, name string) (float64, error) {
	return 0.0, nil
}

func (fakeService) GetCounter(ctx context.Context, name string) (int64, error) {
	return 0, nil
}

func (fakeService) UpdateMetrics(ctx context.Context, input []models.Metrics) error {
	return nil
}

func ExampleNewRouter() {
	h := &Handler{
		service: fakeService{},
		logger:  zap.NewNop(),
	}

	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))

	// Output:
	// 200
	// text/html; charset=utf-8
}
