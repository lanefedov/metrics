package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	models "github.com/lanefedov/metrics/internal/model"
)

func isJSONContentType(contentType string) bool {
	base, _, _ := strings.Cut(contentType, ";")
	return strings.TrimSpace(strings.ToLower(base)) == "application/json"
}

func decodeMetricJSON(body io.Reader) (models.Metrics, error) {
	var metric models.Metrics

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metric); err != nil {
		return models.Metrics{}, err
	}

	var extra struct{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return models.Metrics{}, errors.New("unexpected extra JSON data")
		}

		return models.Metrics{}, err
	}

	return metric, nil
}

func decodeMetricsJSON(body io.Reader) ([]models.Metrics, error) {
	var metrics []models.Metrics

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&metrics); err != nil {
		return nil, err
	}

	var extra struct{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("unexpected extra JSON data")
		}

		return nil, err
	}

	return metrics, nil
}

func writeMetricJSON(w http.ResponseWriter, status int, metric models.Metrics) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(metric)
}

func writeMetricsJSON(w http.ResponseWriter, status int, metrics []models.Metrics) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(metrics)
}
