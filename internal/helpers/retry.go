package helpers

import (
	"context"
	"errors"
	"time"
)

func WithRetry[T any](ctx context.Context, attempts int, isRetriable func(error) bool, fn func() (T, error)) (T, error) {
	var zero T
	if attempts < 1 {
		return zero, errors.New("attempts count must be positive")
	}
	var err error
	currentDelay := 1 * time.Second
	for i := 0; i < attempts; i++ {

		if ctx.Err() != nil {
			return zero, ctx.Err()
		}

		var result T
		result, err = fn()
		if err == nil {
			return result, nil
		}

		if !isRetriable(err) {
			return zero, err
		}

		if i < attempts-1 {
			timer := time.NewTimer(currentDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return zero, ctx.Err()
			case <-timer.C:
			}
			currentDelay += 2 * time.Second
		}
	}
	return zero, err
}
