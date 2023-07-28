package agent

import (
	"context"
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"io"
	"net/http"
	"strings"
	"time"
)

type MetricsAgentConfig interface {
	MetricsReceiverAddress() string
	SendMetricsTimeout() time.Duration
}

type httpMetricsAgent struct {
	client           http.Client
	metricsServerURL string
	pushTimeout      time.Duration
}

func NewHTTPMetricsAgent(config MetricsAgentConfig) (MetricAgent, error) {
	return &httpMetricsAgent{
		client:           http.Client{},
		metricsServerURL: strings.TrimRight(config.MetricsReceiverAddress(), "/"),
		pushTimeout:      config.SendMetricsTimeout(),
	}, nil
}

func (p *httpMetricsAgent) Send(ctx context.Context, metrics []metrics.Metric) error {
	logger.InfoFormat("Push %v metrics", len(metrics))

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	for _, metric := range metrics {
		metricType := metric.GetType()
		metricName := metric.GetName()
		metricValue := metric.GetStringValue()
		defer metric.JumpToTheOriginalState()

		url := fmt.Sprintf("%v/update/%v/%v/%v", p.metricsServerURL, metricType, metricName, metricValue)
		request, err := http.NewRequestWithContext(pushCtx, http.MethodPost, url, nil)
		if err != nil {
			logger.WrapError("Fail to create push request: %v", err).Error()
			return err
		}
		request.Header.Add("Content-Type", "text/plain")

		response, err := p.client.Do(request)
		if err != nil {
			logger.WrapError("Fail to push metric: %v", err).Error()
			return err
		}
		defer response.Body.Close()

		content, err := io.ReadAll(response.Body)
		if err != nil {
			logger.WrapError("Fail to read response body: %v", err).Error()
			return err
		}

		stringContent := string(content)
		if response.StatusCode != http.StatusOK {
			fmt.Printf("Unexpected response status code: %v %v", response.Status, stringContent)
			return fmt.Errorf("fail to push metric: %v", stringContent)
		}

		fmt.Printf("Pushed metric: %v. value: %v, status: %v %v",
			metricName, metric.GetStringValue(), response.Status, stringContent)
	}
	return nil
}
