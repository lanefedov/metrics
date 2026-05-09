package handler

import (
	"html/template"
	"net/http"
	"strconv"

	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/service"
)

var metricsPageTemplate = template.Must(template.New("metrics").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Metrics</title>
</head>
<body>
<ul>
{{range .}}
<li>{{.Type}} {{.Name}}: {{.Value}}</li>
{{end}}
</ul>
</body>
</html>
`))

type metricEntry struct {
	Type  string
	Name  string
	Value string
}

// ListHandler обрабатывает GET /.
type ListHandler struct {
	service *service.MetricsService
}

// NewListHandler создаёт обработчик списка всех метрик.
func NewListHandler(metricsService *service.MetricsService) *ListHandler {
	return &ListHandler{service: metricsService}
}

// ServeHTTP реализует net/http.Handler.
func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.ListMetrics()
	entries := make([]metricEntry, 0, len(metrics))

	for _, metric := range metrics {
		entry := metricEntry{
			Type: metric.MType,
			Name: metric.ID,
		}

		switch metric.MType {
		case models.Counter:
			entry.Value = strconv.FormatInt(*metric.Delta, 10)
		case models.Gauge:
			entry.Value = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		default:
			continue
		}

		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = metricsPageTemplate.Execute(w, entries)
}
