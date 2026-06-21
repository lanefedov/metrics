package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	models "github.com/lanefedov/metrics/internal/model"
	"github.com/lanefedov/metrics/internal/retry"
)

// Reporter отправляет метрики на сервер по HTTP.
type Reporter struct {
	client      *resty.Client
	key         string
	retryDelays []time.Duration
	retrySleep  retry.SleepFunc
}

// NewReporter создаёт HTTP-отправитель метрик на базе resty-клиента.
func NewReporter(baseURL string, key string, client *resty.Client) *Reporter {
	if client == nil {
		client = resty.New()
	}

	return &Reporter{
		client:      client.SetBaseURL(strings.TrimRight(normalizeBaseURL(baseURL), "/")),
		key:         key,
		retryDelays: retry.DefaultDelays,
		retrySleep:  retry.Sleep,
	}
}

// Report отправляет все переданные метрики одним батчем.
func (r *Reporter) Report(metrics []Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	body, err := json.Marshal(metricsToModels(metrics))
	if err != nil {
		return err
	}

	var hashHeader string
	if r.key != "" {
		mac := hmac.New(sha256.New, []byte(r.key))
		mac.Write(body)
		hashHeader = hex.EncodeToString(mac.Sum(nil))
	}

	compressedBody, err := gzipData(body)
	if err != nil {
		return err
	}

	var resp *resty.Response
	if err := retry.DoWithSleeper(
		context.Background(),
		r.retryDelays,
		func(error) bool {
			return true
		},
		func() error {
			var requestErr error
			req := r.client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Content-Encoding", "gzip").
				SetBody(compressedBody)
			if hashHeader != "" {
				req.SetHeader("HashSHA256", hashHeader)
			}
			resp, requestErr = req.Post("/updates/")
			if requestErr != nil {
				return fmt.Errorf("post metrics: %w", requestErr)
			}
			return nil
		},
		r.retrySleep,
	); err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return nil
}

func metricsToModels(metrics []Metric) []models.Metrics {
	requestMetrics := make([]models.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		requestMetrics = append(requestMetrics, metricToModel(metric))
	}

	return requestMetrics
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
