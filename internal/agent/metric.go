package agent

import (
	"strconv"

	models "github.com/lanefedov/metrics/internal/model"
)

// Metric is the internal representation used by the agent.
type Metric struct {
	Name         string
	Type         string
	GaugeValue   float64
	CounterValue int64
}

// PathValue formats the metric value for the HTTP path.
func (m Metric) PathValue() string {
	if m.Type == models.Counter {
		return strconv.FormatInt(m.CounterValue, 10)
	}

	return strconv.FormatFloat(m.GaugeValue, 'f', -1, 64)
}
