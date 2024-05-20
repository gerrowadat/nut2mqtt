package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Nicked form the example at https://pkg.go.dev/github.com/prometheus/client_golang/prometheus

type metrics struct {
	ControlMessagesProcessed    prometheus.Counter
	UPSScrapesCount             prometheus.Counter
	UPSVariableUpdatesProcessed prometheus.Counter
	MQTTUpdatesProcessed        prometheus.Counter
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
		UPSVariableUpdatesProcessed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ups_variable_updates_processed",
				Help: "Number of UPS variable updates processed.",
			},
		),
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

func NewUPSMetrics(reg prometheus.Registerer) map[string]prometheus.Collector {
	u := map[string]prometheus.Collector{}
	for _, v := range UPSMetricsList {
		if v.Type == "gaugevec" {
			u[v.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: v.Name,
				Help: v.Help,
			}, []string{"host", "ups"})
			reg.MustRegister(u[v.Name])
		}
	}
	return u
}

type MetricRegistry struct {
	registry    *prometheus.Registry
	metrics     *metrics
	ups_metrics *map[string]prometheus.Collector
}

func NewMetricRegistry() *MetricRegistry {
	var registry *prometheus.Registry = prometheus.NewRegistry()
	m := NewMetrics(registry)
	u := NewUPSMetrics(registry)
	return &MetricRegistry{registry, m, &u}
}

func (m *MetricRegistry) Metrics() *metrics {
	return m.metrics
}

func (m *MetricRegistry) Registry() *prometheus.Registry {
	return m.registry
}

func (m *MetricRegistry) UPSMetrics() *map[string]prometheus.Collector {
	return m.ups_metrics
}

func (m *MetricRegistry) GetUPSMetric(name string) prometheus.Collector {
	return (*m.ups_metrics)[name]
}

type UPSMetrics struct {
	Name        string
	Help        string
	NutVariable string
	Type        string
}

var UPSMetricsList = []UPSMetrics{
	{
		Name:        "ups_output_voltage",
		Help:        "UPS Output Voltage",
		NutVariable: "output.voltage",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_output_current",
		Help:        "UPS Output Current",
		NutVariable: "output.current",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_output_frequency",
		Help:        "UPS Output frequency",
		NutVariable: "output.frequency",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_input_voltage",
		Help:        "UPS Input Voltage",
		NutVariable: "input.voltage",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_input_current",
		Help:        "UPS Input Current",
		NutVariable: "input.current",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_input_frequency",
		Help:        "UPS Input Frequency",
		NutVariable: "input.frequency",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_load",
		Help:        "UPS Load",
		NutVariable: "ups.load",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_battery_voltage",
		Help:        "UPS Battery Voltage",
		NutVariable: "battery.voltage",
		Type:        "gaugevec",
	},
	{
		Name:        "ups_battery_charge",
		Help:        "Battery charge percentage",
		NutVariable: "battery.charge",
		Type:        "gaugevec",
	},
}
