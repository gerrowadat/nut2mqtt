package http

import (
	"net/http"

	control "github.com/gerrowadat/nut2mqtt/internal/control"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func HTTPServer(c *control.Controller, listen *string) {
	defer c.WaitGroupDone()
	http.HandleFunc("/", RootHandler)
	http.Handle("/metrics", promhttp.HandlerFor(c.MetricRegistry().Registry(), promhttp.HandlerOpts{Registry: c.MetricRegistry().Registry()}))

	http.ListenAndServe(*listen, nil)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<a href=\"/metrics\">/metrics</a>"))
}
