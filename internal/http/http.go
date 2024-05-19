package http

import (
	"net/http"
	"sync"

	control "github.com/gerrowadat/nut2mqtt/internal/control"
	"github.com/gerrowadat/nut2mqtt/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func HTTPServer(controller *control.Controller, listen *string, mr *metrics.MetricRegistry, wg *sync.WaitGroup) {
	defer wg.Done()
	http.HandleFunc("/", RootHandler)
	http.Handle("/metrics", promhttp.HandlerFor(mr.Registry(), promhttp.HandlerOpts{Registry: mr.Registry()}))

	http.ListenAndServe(*listen, nil)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<a href=\"/metrics\">/metrics</a>"))
}
