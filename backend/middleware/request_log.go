// Package middleware reúne protecciones y observabilidad HTTP compartidas.
package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger registra cada respuesta HTTP con datos operativos mínimos.
// No registra cuerpos, cookies, encabezados ni parámetros de consulta para
// evitar que contraseñas, sesiones o información personal terminen en logs.
func RequestLogger(next http.Handler, logger *log.Logger) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		response := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(response, r)

		level := "INFO"
		if response.status >= http.StatusBadRequest {
			level = "WARN"
		}
		if response.status >= http.StatusInternalServerError {
			level = "ERROR"
		}
		logger.Printf("level=%s event=http_request method=%q path=%q status=%d duration=%q", level, r.Method, r.URL.Path, response.status, time.Since(startedAt).Round(time.Microsecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusRecorder) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
