package handler

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/storage"
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

// ValueHandler обрабатывает GET /value/{type}/{name}.
type ValueHandler struct {
	store storage.MetricsReader
}

// NewValueHandler создаёт обработчик чтения метрики.
func NewValueHandler(store storage.MetricsReader) *ValueHandler {
	return &ValueHandler{store: store}
}

// ServeHTTP реализует net/http.Handler.
func (h *ValueHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	var (
		value string
		ok    bool
	)

	switch metricType {
	case models.Gauge:
		var gauge float64
		gauge, ok = h.store.GetGauge(name)
		value = strconv.FormatFloat(gauge, 'f', -1, 64)
	case models.Counter:
		var counter int64
		counter, ok = h.store.GetCounter(name)
		value = strconv.FormatInt(counter, 10)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}

// ListHandler обрабатывает GET /.
type ListHandler struct {
	store storage.MetricsLister
}

// NewListHandler создаёт обработчик списка всех метрик.
func NewListHandler(store storage.MetricsLister) *ListHandler {
	return &ListHandler{store: store}
}

// ServeHTTP реализует net/http.Handler.
func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	entries := make([]metricEntry, 0)

	for name, value := range h.store.ListCounters() {
		entries = append(entries, metricEntry{
			Type:  models.Counter,
			Name:  name,
			Value: strconv.FormatInt(value, 10),
		})
	}

	for name, value := range h.store.ListGauges() {
		entries = append(entries, metricEntry{
			Type:  models.Gauge,
			Name:  name,
			Value: strconv.FormatFloat(value, 'f', -1, 64),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type == entries[j].Type {
			return entries[i].Name < entries[j].Name
		}

		return entries[i].Type < entries[j].Type
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = metricsPageTemplate.Execute(w, entries)
}
