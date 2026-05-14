package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics описывает метрику в плоской структуре без дополнительной вложенности.
// Поля Delta и Value объявлены указателями, чтобы отличать значение `0`
// от отсутствующего значения и не включать его в сериализованную структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
