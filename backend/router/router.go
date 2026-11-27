// Package router define las rutas públicas de FTTH en un único lugar.
package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"ftth/backend/handler"
)

// SetupRoutes registra la API. La lógica de cada operación permanece en handler.
func SetupRoutes(mux *http.ServeMux, server *handler.Server) {
	mux.HandleFunc("POST /api/login", server.Login)
	mux.HandleFunc("POST /api/logout", server.Logout)
	mux.HandleFunc("GET /api/me", server.Me)

	mux.HandleFunc("GET /api/solicitudes", server.ListRequests)
	mux.HandleFunc("POST /api/solicitudes", server.CreateRequest)
	mux.HandleFunc("PUT /api/solicitudes/{id}", server.UpdateRequest)
	mux.HandleFunc("DELETE /api/solicitudes/{id}", server.DeleteRequest)

	mux.HandleFunc("GET /api/modulos/{module}", server.ListModuleItems)
	mux.HandleFunc("POST /api/modulos/{module}", server.CreateModuleItem)
	mux.HandleFunc("PUT /api/modulos/{module}/{id}", server.UpdateModuleItem)
	mux.HandleFunc("DELETE /api/modulos/{module}/{id}", server.DeleteModuleItem)
}

// SPAHandler publica la interfaz compilada y deriva las rutas de cliente al HTML
// principal, sin interferir con las rutas /api definidas arriba.
func SPAHandler(files http.FileSystem) http.Handler {
	fileServer := http.FileServer(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && !strings.Contains(filepath.Base(r.URL.Path), ".") {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
