package audit

import (
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

	if v.client != nil {
		v.client = nil
	}
	v.auditURL = ""
}
