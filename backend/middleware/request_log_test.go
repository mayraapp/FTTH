package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLoggerRecordsMethodPathStatusAndNotQuery(t *testing.T) {
	var output bytes.Buffer
	logger := log.New(&output, "", 0)
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}), logger)

	request := httptest.NewRequest(http.MethodGet, "/api/me?token=secret-value", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("estado esperado 401, recibido %d", response.Code)
	}
	entry := output.String()
	for _, expected := range []string{`level=WARN`, `method="GET"`, `path="/api/me"`, `status=401`, `duration=`} {
		if !strings.Contains(entry, expected) {
			t.Fatalf("falta %q en el registro: %s", expected, entry)
		}
	}
	if strings.Contains(entry, "secret-value") || strings.Contains(entry, "token=") {
		t.Fatalf("el registro no debe incluir parámetros de consulta: %s", entry)
	}
}
