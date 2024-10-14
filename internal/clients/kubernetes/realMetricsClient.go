package kubernetes

import (
	"context"

	ultron "github.com/be-heroes/ultron/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type RealMetricsClient struct {
	ClientSet *metricsclient.Clientset
}

func (r *RealMetricsClient) ListNodeMetrics(ctx context.Context, opts metav1.ListOptions) (*ultron.MetricsNodeList, error) {
	nodeMetricsList, err := r.ClientSet.MetricsV1beta1().NodeMetricses().List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var metricsNodeList ultron.MetricsNodeList
	for _, item := range nodeMetricsList.Items {
		metricsNodeList.Items = append(metricsNodeList.Items, ultron.MetricsNode{
			Name:  item.Name,
			Usage: item.Usage,
		})
	}

	return &metricsNodeList, nil
}

func (r *RealMetricsClient) ListPodMetrics(ctx context.Context, namespace string, opts metav1.ListOptions) (*ultron.MetricsPodList, error) {
	podMetricsList, err := r.ClientSet.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var metricsPodList ultron.MetricsPodList
	for _, item := range podMetricsList.Items {
		var containers []ultron.MetricsContainer
		for _, c := range item.Containers {
			containers = append(containers, ultron.MetricsContainer{
				Name:  c.Name,
				Usage: c.Usage,
			})
		}
		metricsPodList.Items = append(metricsPodList.Items, ultron.MetricsPod{
			Name:       item.Name,
			Namespace:  item.Namespace,
			Containers: containers,
		})
	}

	return &metricsPodList, nil
}
