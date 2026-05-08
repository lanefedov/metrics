package agent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Reporter sends metrics to the server over HTTP.
type Reporter struct {
	baseURL string
	client  httpDoer
}

// NewReporter creates an HTTP reporter backed by http.Client.
func NewReporter(baseURL string, client httpDoer) *Reporter {
	if client == nil {
		client = http.DefaultClient
	}

	return &Reporter{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

// Report sends all provided metrics one-by-one.
func (r *Reporter) Report(metrics []Metric) error {
	var reportErr error

	for _, metric := range metrics {
		if err := r.sendMetric(metric); err != nil {
			reportErr = err
		}
	}

	return reportErr
}

func (r *Reporter) sendMetric(metric Metric) error {
	path := fmt.Sprintf(
		"%s/update/%s/%s/%s",
		r.baseURL,
		metric.Type,
		url.PathEscape(metric.Name),
		url.PathEscape(metric.PathValue()),
	)

	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

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
