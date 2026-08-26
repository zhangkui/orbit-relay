package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Envelope struct {
	OK        bool      `json:"ok"`
	Data      any       `json:"data,omitempty"`
	Error     *Problem  `json:"error,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	At        time.Time `json:"at"`
}
type Problem struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{OK: status < 400, Data: data, At: time.Now().UTC()})
}
func Fail(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{OK: false, Error: &Problem{Code: code, Message: message}, At: time.Now().UTC()})
}
func Decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func Method(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	for _, x := range allowed {
		if r.Method == x {
			return true
		}
	}
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	Fail(w, http.StatusMethodNotAllowed, "method_not_allowed", "HTTP method is not allowed")
	return false
}
