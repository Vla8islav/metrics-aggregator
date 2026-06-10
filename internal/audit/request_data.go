package audit

import "context"

type requestDataKey struct{}

type RequestData struct {
	Metrics []string
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
