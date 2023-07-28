package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/model"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/sendler"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"net/url"
	"time"
)

type metricsPusherConfig interface {
	ParallelLimit() int
	MetricsServerURL() string
	PushMetricsTimeout() time.Duration
}

type httpMetricsPusher struct {
	parallelLimit    int
	client           http.Client
	metricsServerURL string
	pushTimeout      time.Duration
	converter        *model.MetricsConverter
}

func NewMetricsPusher(config metricsPusherConfig, converter *model.MetricsConverter) (sendler.MetricsPusher, error) {
	serverURL, err := normalizeURL(config.MetricsServerURL())
	if err != nil {
		return nil, logger.WrapError("normalize url", err)
	}

	return &httpMetricsPusher{
		parallelLimit:    config.ParallelLimit(),
		client:           http.Client{},
		metricsServerURL: serverURL.String(),
		pushTimeout:      config.PushMetricsTimeout(),
		converter:        converter,
	}, nil
}

func (p *httpMetricsPusher) Push(ctx context.Context, metricsChan <-chan metrics.Metric) error {
	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < p.parallelLimit; i++ {
		eg.Go(func() error {
			for {
				select {
				case metric, ok := <-metricsChan:
					if !ok {
						return nil
					}

					err := p.pushMetrics(ctx, []metrics.Metric{metric})
					if err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	return eg.Wait()
}

func (p *httpMetricsPusher) pushMetrics(ctx context.Context, metricsList []metrics.Metric) error {
	metricsCount := len(metricsList)
	if metricsCount == 0 {
		logger.Info("Nothing to push")
	}
	logger.InfoFormat("Push %v metrics", metricsCount)

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	modelMetrics := make([]*model.Metrics, metricsCount)
	for i, metric := range metricsList {
		modelMetric, err := p.converter.ToModelMetric(metric)
		if err != nil {
			return logger.WrapError("create model request", err)
		}

		modelMetrics[i] = modelMetric
	}

	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(modelMetrics)
	if err != nil {
		return logger.WrapError("serialize model request", err)
	}

	request, err := http.NewRequestWithContext(pushCtx, http.MethodPost, p.metricsServerURL+"/updates", &buffer)
	if err != nil {
		return logger.WrapError("create push request", err)
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := p.client.Do(request)
	if err != nil {
		return logger.WrapError("push metrics", err)
	}

	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return logger.WrapError("read response body", err)
	}

	stringContent := string(content)
	if response.StatusCode != http.StatusOK {
		logger.ErrorFormat("Unexpected response status code: %v %v", response.Status, stringContent)
		return logger.WrapError(fmt.Sprintf("push metric: %s", stringContent), metrics.ErrUnexpectedStatusCode)
	}

	for _, metric := range metricsList {
		logger.InfoFormat("Pushed metric: %v. value: %v, status: %v", metric.GetName(), metric.GetStringValue(), response.Status)
		metric.JumpToTheOriginalState()
	}

	return nil
}

func normalizeURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, logger.WrapError("normalize url", metrics.ErrEmptyURL)
	}

	result, err := url.ParseRequestURI(urlStr)
	if err != nil {
		result, err = url.ParseRequestURI("http://" + urlStr)
		if err != nil {
			return nil, logger.WrapError("parse request url", err)
		}
	}

	if result.Scheme == "localhost" {
		// =)
		return normalizeURL("http://" + result.String())
	}

	if result.Scheme == "" {
		result.Scheme = "http"
	}

	return result, nil
}
