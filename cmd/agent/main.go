package main

import "time"

type config struct {
	metricReceiverAddress  string
	sendTimeout            time.Duration
	pollInterval           time.Duration
	reportInterval         time.Duration
	listOfCollectedMetrics []string
}

func (c *config) GetListOfCollectedMetrics() []string {
	return c.listOfCollectedMetrics
}

func (c *config) GetMetricReceiverAddress() string {
	return c.metricReceiverAddress
}

func (c *config) GetSendTimeout() time.Duration {
	return c.sendTimeout
}
