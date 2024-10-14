package kubernetes

import (
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type MetricsClientWrapper struct {
	ClientSet *metricsclient.Clientset
}

func (m *MetricsClientWrapper) MetricsV1beta1() metricsv1beta1.MetricsV1beta1Interface {
	return m.ClientSet.MetricsV1beta1()
}
