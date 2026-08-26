package handler

import (
	"encoding/json"
	"errors"
	"gitlab.com/zhangkui/orbit-relay/internal/model"
	"gitlab.com/zhangkui/orbit-relay/internal/service"
	"net/http"
	"strings"
)

type Handler struct{ app *service.Lab }

func New(app *service.Lab) http.Handler {
	h := &Handler{app: app}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/api/v1/satellites", h.satellites)
	mux.HandleFunc("/api/v1/stations", h.stations)
	mux.HandleFunc("/api/v1/windows", h.windows)
	mux.HandleFunc("/api/v1/telemetry", h.telemetry)
	mux.HandleFunc("/api/v1/commands", h.commands)
	mux.HandleFunc("/api/v1/reports/", h.report)
	mux.HandleFunc("/", h.index)
	return logging(mux)
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]string{"status": "ok", "service": "orbit-relay"})
}
func (h *Handler) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>Orbit Relay</title><style>body{font:16px system-ui;background:#0b1320;color:#dbeafe;margin:40px}h1{color:#7dd3fc}section{border:1px solid #28435e;padding:20px;margin:16px 0;border-radius:8px}</style></head><body><h1>Orbit Relay Ground Station</h1><section>卫星地面站通信窗口控制服务</section><section id="health">正在检查链路...</section><script>fetch('/healthz').then(r=>r.json()).then(x=>document.querySelector('#health').textContent='服务状态：'+x.status)</script></body></html>`))
}
func (h *Handler) satellites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var s model.Satellite
	if decode(r, &s) != nil {
		write(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.app.AddSatellite(r.Context(), s); err != nil {
		writeErr(w, err)
		return
	}
	write(w, 201, s)
}
func (h *Handler) stations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var s model.GroundStation
	if decode(r, &s) != nil {
		write(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.app.AddStation(r.Context(), s); err != nil {
		writeErr(w, err)
		return
	}
	write(w, 201, s)
}
func (h *Handler) windows(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var x model.ContactWindow
		if decode(r, &x) != nil {
			write(w, 400, map[string]string{"error": "invalid json"})
			return
		}
		if err := h.app.PlanWindow(r.Context(), x); err != nil {
			writeErr(w, err)
			return
		}
		write(w, 201, x)
		return
	}
	http.Error(w, "method not allowed", 405)
}
func (h *Handler) telemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var f model.TelemetryFrame
	if decode(r, &f) != nil {
		write(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.app.Ingest(r.Context(), f); err != nil {
		writeErr(w, err)
		return
	}
	write(w, 202, map[string]string{"status": "accepted", "id": f.ID})
}
func (h *Handler) commands(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, 200, h.app.ListCommands())
		return
	}
	if r.Method == http.MethodPost {
		var p model.CommandPacket
		if decode(r, &p) != nil {
			write(w, 400, map[string]string{"error": "invalid json"})
			return
		}
		if err := h.app.QueueCommand(r.Context(), p); err != nil {
			writeErr(w, err)
			return
		}
		write(w, 202, p)
		return
	}
	http.Error(w, "method not allowed", 405)
}
func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/reports/")
	x, err := h.app.Report(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	write(w, 200, x)
}
func decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func write(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, err error) {
	status := 500
	if errors.Is(err, model.ErrInvalid) {
		status = 400
	}
	if errors.Is(err, model.ErrNotFound) {
		status = 404
	}
	if errors.Is(err, model.ErrConflict) {
		status = 409
	}
	if errors.Is(err, model.ErrWindowClosed) {
		status = 422
	}
	write(w, status, map[string]string{"error": err.Error()})
}
