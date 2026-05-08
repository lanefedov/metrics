package agent

import (
	"strconv"

	models "github.com/lanefedov/metrics/internal/model"
)

// Metric описывает внутреннее представление метрики в агенте.
type Metric struct {
	Name         string
	Type         string
	GaugeValue   float64
	CounterValue int64
}

// PathValue форматирует значение метрики для HTTP-пути.
func (m Metric) PathValue() string {
	if m.Type == models.Counter {
		return strconv.FormatInt(m.CounterValue, 10)
	}

	return strconv.FormatFloat(m.GaugeValue, 'f', -1, 64)
}
