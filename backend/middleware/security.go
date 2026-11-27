// Package middleware reúne las protecciones HTTP compartidas por la aplicación.
package middleware

import "net/http"

// SecurityHeaders aplica cabeceras seguras sin mezclar esa responsabilidad con
// las rutas ni con los handlers de negocio.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}
