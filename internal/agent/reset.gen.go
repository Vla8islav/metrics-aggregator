package agent

import (
	"time"
)

func resetPointer[T any](v *T) {
	if v == nil {
		return
	}

	if resetter, ok := any(v).(interface{ Reset() }); ok {
		resetter.Reset()
		return
	}

	*v = *new(T)
}

func (v *Agent) Reset() {
	if v == nil {
		return
	}

	resetPointer(v.client)
	v.serverAddr = ""
	v.pollInterval = *new(time.Duration)
	v.reportInterval = *new(time.Duration)
	v.rateLimit = 0
	resetPointer(v.gauges)
	resetPointer(v.config)
	resetPointer(v.logger)
}
