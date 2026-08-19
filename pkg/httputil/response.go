// Package httputil provides the response envelope every REST response
// (success or error) is wrapped in, so gRPC-Gateway handlers never need to
// think about response shape themselves.
package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/telemetry"
)

// Envelope is the shape every REST response takes.
type Envelope struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type bufferingWriter struct {
	http.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (w *bufferingWriter) WriteHeader(status int)      { w.status = status }
func (w *bufferingWriter) Write(b []byte) (int, error) { return w.buf.Write(b) }

// ResponseWrapperMiddleware buffers the downstream handler's JSON body and
// re-emits it wrapped in Envelope, attaching the request's trace ID. Errors
// (grpc-gateway maps non-2xx statuses through its own error marshaler,
// which already produces JSON with a "message"/"code" field) are passed
// through under Error instead of Data.
func ResponseWrapperMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bw := &bufferingWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(bw, r)

			envelope := Envelope{
				Success:   bw.status < 400,
				Message:   http.StatusText(bw.status),
				TraceID:   telemetry.TraceID(r.Context()),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			var payload interface{}
			if bw.buf.Len() > 0 {
				if err := json.Unmarshal(bw.buf.Bytes(), &payload); err != nil {
					payload = bw.buf.String()
				}
			}
			if envelope.Success {
				envelope.Data = payload
			} else {
				envelope.Error = payload
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(bw.status)
			if err := json.NewEncoder(w).Encode(envelope); err != nil {
				log.WithContext(r.Context()).Error("failed to encode response envelope")
			}
		})
	}
}
