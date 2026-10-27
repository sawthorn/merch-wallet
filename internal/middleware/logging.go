// Package middleware provides middleware functions
// used for wrapping http Handlers
package middleware

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/felixge/httpsnoop"
)

func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)
		log.Printf("%s %s (code=%d dur=%s)", r.Method, r.URL, m.Code, m.Duration)
	})
}
