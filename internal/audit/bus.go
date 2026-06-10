package audit

import (
	"context"
	"time"
)

type Event struct {
	Time       time.Time `json:"ts"`
	Metrics    []string  `json:"metrics"`
	RemoteAddr string    `json:"ip_address"`
	Operation  string    `json:"operation"`
}

type Sink interface {
	Write(ctx context.Context, event Event) error
}

type Publisher struct {
	sinks []Sink
}

func NewPublisher(sinks ...Sink) *Publisher {
	return &Publisher{sinks: sinks}
}

func (p *Publisher) Publish(ctx context.Context, event Event) error {
	for _, sink := range p.sinks {
		if err := sink.Write(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
