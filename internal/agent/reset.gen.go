package agent

import (
	"time"
)

func (v *Agent) Reset() {
	if v == nil {
		return
	}

	if v.client != nil {
		v.client = nil
	}
	v.serverAddr = ""
	v.pollInterval = *new(time.Duration)
	v.reportInterval = *new(time.Duration)
	v.rateLimit = 0
	if v.gauges != nil {
		v.gauges = nil
	}
	if v.config != nil {
		v.config = nil
	}
	if v.logger != nil {
		v.logger = nil
	}
}
