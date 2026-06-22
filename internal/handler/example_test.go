package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"go.uber.org/zap"
)

type fakeService struct{}

func (fakeService) GetAll(ctx context.Context) (models.MetricsExport, error) {
	return models.MetricsExport{
		Gauges: map[string]float64{
			"Alloc":       1.124,
			"RandomValue": 42,
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

	for _, request := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/"},
		{method: http.MethodGet, path: "/ping"},
		{method: http.MethodGet, path: "/value/gauge/Alloc"},
		{method: http.MethodPost, path: "/update/counter/PollCount/1"},
	} {
		req := httptest.NewRequest(request.method, request.path, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		fmt.Println(request.method, request.path, w.Code)
		if request.path == "/" {
			fmt.Println(strings.Contains(w.Body.String(), "<h1>Metrics</h1>"))
			fmt.Println(strings.Contains(w.Body.String(), "<strong>Alloc</strong>: 1.124"))
			fmt.Println(strings.Contains(w.Body.String(), "<strong>RandomValue</strong>: 42"))
			fmt.Println(strings.Contains(w.Body.String(), "<strong>PollCount</strong>: 10"))
		}
	}

	// Output:
	// GET / 200
	// true
	// true
	// true
	// true
	// GET /ping 200
	// GET /value/gauge/Alloc 200
	// POST /update/counter/PollCount/1 200
}

func ExampleNewRouter_updateBatchMetrics() {
	h := &Handler{
		service: fakeService{},
		logger:  zap.NewNop(),
	}

	router := NewRouter(h)

	body := strings.NewReader(`[
		{"id":"Alloc","type":"gauge","value":123.45},
		{"id":"PollCount","type":"counter","delta":10}
	]`)

	req := httptest.NewRequest(http.MethodPost, "/updates", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Println(w.Code)

	// Output:
	// 200
}
