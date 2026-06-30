package agent

import (
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"github.com/Vla8islav/metrics-aggregator/internal/model"
	"go.uber.org/zap"
)

func (v *Agent) Reset() {
	if v == nil {
		return
	}

	v.client = new(helpers.HTTPRetryClient)
	v.serverAddr = ""
	v.pollInterval = *new(time.Duration)
	v.reportInterval = *new(time.Duration)
	v.rateLimit = 0
	v.gauges = new(models.Stats)
	v.config = new(config.OptionsClient)
	v.logger = new(zap.Logger)
}
