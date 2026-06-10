package audit

import (
	"context"
	"time"
)

type Event struct {
	Time        time.Time `json:"ts"`
	MetricsList []string  `json:"metrics"`
	RemoteAddr  string    `json:"ip_address"`
}

type Sink interface {
	Write(ctx context.Context, event Event) error
}

type Publisher struct {
	sinks []Sink
}

func NewPublisher(sinks ...Sink) *Publisher {
	return &Publisher{sinks}
}

func (p *Publisher) Publish(ctx context.Context, event Event) error {
	for _, sink := range p.sinks {
		if err := sink.Write(ctx, event); err != nil {
			// log and continue; audit failure should not break request handling
		}
	}
	return nil
}
