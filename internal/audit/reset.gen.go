package audit

import (
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"time"
)


func (v *Event) Reset() {
	if v == nil {
		return
	}

	v.Time = *new(UnixTime)
	v.Metrics = v.Metrics[:0]
	v.RemoteAddr = ""
	v.Operation = ""
}


func (v *UnixTime) Reset() {
	if v == nil {
		return
	}

	v.Time = *new(time.Time)
}


func (v *WebSink) Reset() {
	if v == nil {
		return
	}

	v.client = new(helpers.HTTPRetryClient)
	v.auditURL = ""
}

