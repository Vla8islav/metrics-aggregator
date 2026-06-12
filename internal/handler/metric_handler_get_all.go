package handler

import (
	"html/template"
	"log"
	"net/http"

	"go.uber.org/zap"
)

var metricsPageTemplate = template.Must(template.New("metrics").Parse(`
<html>
<body>
    <h1>Metrics</h1>

    <h2>Gauges</h2>
    {{if .Gauges}}
    <ul>
        {{range $name, $value := .Gauges}}
        <li><strong>{{$name}}</strong>: {{$value}}</li>
        {{end}}
    </ul>
    {{else}}
    <p>No gauges available.</p>
    {{end}}

    <h2>Counters</h2>
    {{if .Counters}}
    <ul>
        {{range $name, $value := .Counters}}
        <li><strong>{{$name}}</strong>: {{$value}}</li>
        {{end}}
    </ul>
    {{else}}
    <p>No counters available.</p>
    {{end}}
</body>
</html>`))

// GetAllMetrics prints a simple HTML with all of the exported metrics
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeMethodNotAllowed(w, "only GET method is allowed")
		return
	}

	metricsExport, err := h.service.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all metrics: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := metricsPageTemplate.Execute(w, metricsExport); err != nil {
		h.logger.Error("error rendering metrics template", zap.Error(err))
	}

}
