package audit

import "context"

type requestDataKey struct{}

type RequestData struct {
	Operation string
	Metrics   []string
}

func NewRequestData() *RequestData {
	return &RequestData{
		Metrics: make([]string, 0),
	}
}

func WithRequestData(ctx context.Context, data *RequestData) context.Context {
	return context.WithValue(ctx, requestDataKey{}, data)
}

func FromContext(ctx context.Context) *RequestData {
	data, _ := ctx.Value(requestDataKey{}).(*RequestData)
	return data
}

func AddMetric(ctx context.Context, name string) {
	if name == "" {
		return
	}

	data := FromContext(ctx)
	if data == nil {
		return
	}

	data.Metrics = append(data.Metrics, name)
}

func SetOperation(ctx context.Context, operation string) {
	if operation == "" {
		return
	}

	data := FromContext(ctx)
	if data == nil {
		return
	}

	data.Operation = operation
}
