package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	models "github.com/lanefedov/metrics/internal/model"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Reporter отправляет метрики на сервер по HTTP.
type Reporter struct {
	baseURL string
	client  httpDoer
}

// NewReporter создаёт HTTP-отправитель метрик на базе `http.Client`.
func NewReporter(baseURL string, client httpDoer) *Reporter {
	if client == nil {
		client = http.DefaultClient
	}

	return &Reporter{
		baseURL: strings.TrimRight(normalizeBaseURL(baseURL), "/"),
		client:  client,
	}
}

// Report поочерёдно отправляет все переданные метрики.
func (r *Reporter) Report(metrics []Metric) error {
	var reportErr error

	for _, metric := range metrics {
		if err := r.sendMetric(metric); err != nil {
			reportErr = errors.Join(reportErr, err)
		}
	}

	return reportErr
}

func (r *Reporter) sendMetric(metric Metric) error {
	body, err := json.Marshal(metricToModel(metric))
	if err != nil {
		return err
	}

	compressedBody, err := gzipData(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/update/", r.baseURL), bytes.NewReader(compressedBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func metricToModel(metric Metric) models.Metrics {
	requestMetric := models.Metrics{
		ID:    metric.Name,
		MType: metric.Type,
	}

	switch metric.Type {
	case models.Counter:
		delta := metric.CounterValue
		requestMetric.Delta = &delta
	case models.Gauge:
		value := metric.GaugeValue
		requestMetric.Value = &value
	}

	return requestMetric
}

func normalizeBaseURL(address string) string {
	if strings.Contains(address, "://") {
		return address
	}

	return "http://" + address
}

func gzipData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err := gzipWriter.Write(data); err != nil {
		_ = gzipWriter.Close()
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
