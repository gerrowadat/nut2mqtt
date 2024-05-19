package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Nicked form the example at https://pkg.go.dev/github.com/prometheus/client_golang/prometheus

type metrics struct {
	ControlMessagesProcessed prometheus.Counter
	UPSScrapesCount          prometheus.Counter
	MQTTUpdatesProcessed     prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		ControlMessagesProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "control_messages_processed",
			Help: "Number of control messages processed.",
		}),
		UPSScrapesCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ups_scrapes_count",
			Help: "Number of UPS scrapes performed.",
		}),
		MQTTUpdatesProcessed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "mqtt_updates_processed",
				Help: "Number of MQTT updates processed.",
			},
		),
	}
	reg.MustRegister(m.ControlMessagesProcessed)
	reg.MustRegister(m.UPSScrapesCount)
	reg.MustRegister(m.MQTTUpdatesProcessed)
	return m
}

type MetricRegistry struct {
	registry *prometheus.Registry
	metrics  *metrics
}

func NewMetricRegistry() *MetricRegistry {
	var registry *prometheus.Registry = prometheus.NewRegistry()
	m := NewMetrics(registry)
	return &MetricRegistry{registry, m}
}

func (m *MetricRegistry) Metrics() *metrics {
	return m.metrics
}

func (m *MetricRegistry) Registry() *prometheus.Registry {
	return m.registry
}
