package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ftth/backend/config"
	"ftth/backend/domain"
	"ftth/backend/repository"
	"ftth/backend/session"
)

const (
	sessionName = "ftth_session"
	sessionTTL  = 8 * time.Hour
)

var moduleStatuses = domain.ModuleStatuses

type user = domain.User
type loginInput = domain.LoginInput
type requestInput = domain.RequestInput
type updateRequestInput = domain.UpdateRequestInput
type request = domain.Request
type moduleItemInput = domain.ModuleItemInput
type moduleItem = domain.ModuleItem

// Server contiene el estado de la sesión y los handlers de la API FTTH.
// Las rutas se registran desde backend/router para mantener cada capa aislada.
type Server struct {
	mu             sync.RWMutex
	requests       []request
	modules        map[string][]moduleItem
	sessions       map[string]user
	requestsFile   string
	moduleDataFile string
	dataStore      *repository.PostgresStore
	sessionStore   session.Store
}

// app se conserva como alias privado para las pruebas unitarias del paquete.
type app = Server

// New construye el servidor y carga el estado persistido antes de aceptar rutas.
func New(appConfig config.Config) (*Server, error) {
	a := &Server{
		sessions:       make(map[string]user),
		requestsFile:   appConfig.RequestsFile,
		moduleDataFile: appConfig.ModuleDataFile,
	}
	if err := a.load(); err != nil {
		return nil, fmt.Errorf("no se pudieron cargar las solicitudes: %w", err)
	}
	if appConfig.DatabaseURL != "" {
		store, err := repository.NewPostgresStore(context.Background(), appConfig.DatabaseURL)
		if err != nil {
			return nil, err
		}
		if err := store.Migrate(context.Background()); err != nil {
			store.Close()
			return nil, fmt.Errorf("no se pudo preparar PostgreSQL: %w", err)
		}
		if err := store.SeedIfEmpty(context.Background(), a.requests, a.modules); err != nil {
			store.Close()
			return nil, fmt.Errorf("no se pudo migrar el contenido inicial a PostgreSQL: %w", err)
		}
		a.dataStore = store
		log.Printf("event=storage_ready provider=postgres")
	} else {
		log.Printf("event=storage_ready provider=json")
	}
	if appConfig.RedisAddr != "" {
		store, err := session.NewRedisStore(context.Background(), appConfig.RedisAddr, appConfig.RedisPassword)
		if err != nil {
			if a.dataStore != nil {
				a.dataStore.Close()
			}
			return nil, err
		}
		a.sessionStore = store
		log.Printf("event=session_store_ready provider=redis")
	} else {
		log.Printf("event=session_store_ready provider=memory")
	}
	return a, nil
}

// Métodos exportados para backend/router. Las versiones privadas siguen siendo
// útiles en las pruebas del paquete handler.
func (a *app) Login(w http.ResponseWriter, r *http.Request)            { a.login(w, r) }
func (a *app) Logout(w http.ResponseWriter, r *http.Request)           { a.logout(w, r) }
func (a *app) Me(w http.ResponseWriter, r *http.Request)               { a.me(w, r) }
func (a *app) ListRequests(w http.ResponseWriter, r *http.Request)     { a.listRequests(w, r) }
func (a *app) CreateRequest(w http.ResponseWriter, r *http.Request)    { a.createRequest(w, r) }
func (a *app) UpdateRequest(w http.ResponseWriter, r *http.Request)    { a.updateRequest(w, r) }
func (a *app) DeleteRequest(w http.ResponseWriter, r *http.Request)    { a.deleteRequest(w, r) }
func (a *app) ListModuleItems(w http.ResponseWriter, r *http.Request)  { a.listModuleItems(w, r) }
func (a *app) CreateModuleItem(w http.ResponseWriter, r *http.Request) { a.createModuleItem(w, r) }
func (a *app) UpdateModuleItem(w http.ResponseWriter, r *http.Request) { a.updateModuleItem(w, r) }
func (a *app) DeleteModuleItem(w http.ResponseWriter, r *http.Request) { a.deleteModuleItem(w, r) }

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de acceso inválidos")
		return
	}
	if input.Company != "TERACOM" || input.Username != "admin" || input.Password != "FTTH2026!" {
		writeError(w, http.StatusUnauthorized, "Empresa, usuario o contraseña incorrectos")
		return
	}

	id, err := randomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No fue posible iniciar sesión")
		return
	}
	account := user{Name: "May Torres", Role: "Administrador", Company: "TERACOM"}
	if a.sessionStore != nil {
		if err := a.sessionStore.Create(r.Context(), id, account, sessionTTL); err != nil {
			writeError(w, http.StatusServiceUnavailable, "No fue posible iniciar sesión")
			return
		}
	} else {
		a.mu.Lock()
		a.sessions[id] = account
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionName, Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: int(sessionTTL.Seconds())})
	writeJSON(w, http.StatusOK, map[string]user{"user": account})
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionName); err == nil {
		if a.sessionStore != nil {
			if err := a.sessionStore.Delete(r.Context(), cookie.Value); err != nil {
				writeError(w, http.StatusServiceUnavailable, "No fue posible cerrar la sesión")
				return
			}
		} else {
			a.mu.Lock()
			delete(a.sessions, cookie.Value)
			a.mu.Unlock()
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) me(w http.ResponseWriter, r *http.Request) {
	account, ok := a.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]user{"user": account})
}

func (a *app) listRequests(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	if a.dataStore != nil {
		items, err := a.dataStore.ListRequests(r.Context())
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudieron consultar las solicitudes")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"requests": items})
		return
	}
	a.mu.RLock()
	items := make([]request, len(a.requests))
	copy(items, a.requests)
	a.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	writeJSON(w, http.StatusOK, map[string]any{"requests": items})
}

func (a *app) createRequest(w http.ResponseWriter, r *http.Request) {
	account, ok := a.currentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	var input requestInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de solicitud inválidos")
		return
	}
	input.Type = strings.TrimSpace(input.Type)
	input.Priority = strings.TrimSpace(input.Priority)
	input.Crew = strings.TrimSpace(input.Crew)
	input.Description = strings.TrimSpace(input.Description)
	if input.Priority == "" {
		input.Priority = "Media"
	}
	if input.Crew == "" {
		input.Crew = "Por asignar"
	}
	if input.Type == "" || !validCrew(input.Crew) || len(input.Description) < 10 || len(input.Description) > 600 {
		writeError(w, http.StatusBadRequest, "Completa tipo y una descripción de 10 a 600 caracteres")
		return
	}
	if a.dataStore != nil {
		item, err := a.dataStore.CreateRequest(r.Context(), input, account.Name)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo guardar la solicitud")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]request{"request": item})
		return
	}

	a.mu.Lock()
	nextID := 1
	for _, item := range a.requests {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	item := request{ID: nextID, Folio: fmt.Sprintf("FTTH-%04d", nextID), Type: input.Type, Priority: input.Priority, Crew: input.Crew, Description: input.Description, Status: "Pendiente", CreatedAt: time.Now().UTC(), CreatedBy: account.Name}
	a.requests = append(a.requests, item)
	err := a.saveLocked()
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo guardar la solicitud")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]request{"request": item})
}

func (a *app) updateRequest(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	id, ok := requestID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Identificador de solicitud inválido")
		return
	}
	var input updateRequestInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de solicitud inválidos")
		return
	}
	input.Type = strings.TrimSpace(input.Type)
	input.Priority = strings.TrimSpace(input.Priority)
	input.Crew = strings.TrimSpace(input.Crew)
	input.Description = strings.TrimSpace(input.Description)
	input.Status = strings.TrimSpace(input.Status)
	if input.Type == "" || input.Priority == "" || !validCrew(input.Crew) || len(input.Description) < 10 || len(input.Description) > 600 || !validStatus(input.Status) {
		writeError(w, http.StatusBadRequest, "Completa tipo, prioridad, cuadrilla, estado y una descripción de 10 a 600 caracteres")
		return
	}
	if a.dataStore != nil {
		updated, err := a.dataStore.UpdateRequest(r.Context(), id, input)
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "Solicitud no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo actualizar la solicitud")
			return
		}
		writeJSON(w, http.StatusOK, map[string]request{"request": updated})
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for index, item := range a.requests {
		if item.ID != id {
			continue
		}
		updated := item
		updated.Type = input.Type
		updated.Priority = input.Priority
		updated.Crew = input.Crew
		updated.Description = input.Description
		updated.Status = input.Status
		a.requests[index] = updated
		if err := a.saveLocked(); err != nil {
			a.requests[index] = item
			writeError(w, http.StatusInternalServerError, "No se pudo actualizar la solicitud")
			return
		}
		writeJSON(w, http.StatusOK, map[string]request{"request": updated})
		return
	}
	writeError(w, http.StatusNotFound, "Solicitud no encontrada")
}

func (a *app) deleteRequest(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	id, ok := requestID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Identificador de solicitud inválido")
		return
	}
	if a.dataStore != nil {
		err := a.dataStore.DeleteRequest(r.Context(), id)
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "Solicitud no encontrada")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo eliminar la solicitud")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for index, item := range a.requests {
		if item.ID != id {
			continue
		}
		original := a.requests
		updated := make([]request, 0, len(original)-1)
		updated = append(updated, original[:index]...)
		updated = append(updated, original[index+1:]...)
		a.requests = updated
		if err := a.saveLocked(); err != nil {
			a.requests = original
			writeError(w, http.StatusInternalServerError, "No se pudo eliminar la solicitud")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeError(w, http.StatusNotFound, "Solicitud no encontrada")
}

func (a *app) listModuleItems(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	module, ok := moduleName(r)
	if !ok {
		writeError(w, http.StatusNotFound, "Módulo no encontrado")
		return
	}
	if a.dataStore != nil {
		items, err := a.dataStore.ListModuleItems(r.Context(), module)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudieron consultar los registros")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
		return
	}
	a.mu.RLock()
	items := make([]moduleItem, len(a.modules[module]))
	copy(items, a.modules[module])
	a.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *app) createModuleItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	module, ok := moduleName(r)
	if !ok {
		writeError(w, http.StatusNotFound, "Módulo no encontrado")
		return
	}
	input, ok := validModuleInput(w, r, module)
	if !ok {
		return
	}
	if a.dataStore != nil {
		item, err := a.dataStore.CreateModuleItem(r.Context(), module, input)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo guardar el registro")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]moduleItem{"item": item})
		return
	}

	a.mu.Lock()
	nextID := 1
	for _, item := range a.modules[module] {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	now := time.Now().UTC()
	item := moduleItem{ID: nextID, Name: input.Name, Detail: input.Detail, Status: input.Status, Map: input.Map, CreatedAt: now, UpdatedAt: now}
	a.modules[module] = append(a.modules[module], item)
	err := a.saveModulesLocked()
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo guardar el registro")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]moduleItem{"item": item})
}

func (a *app) updateModuleItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	module, ok := moduleName(r)
	if !ok {
		writeError(w, http.StatusNotFound, "Módulo no encontrado")
		return
	}
	id, ok := moduleItemID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Identificador de registro inválido")
		return
	}
	input, ok := validModuleInput(w, r, module)
	if !ok {
		return
	}
	if a.dataStore != nil {
		updated, err := a.dataStore.UpdateModuleItem(r.Context(), module, id, input)
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "Registro no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo actualizar el registro")
			return
		}
		writeJSON(w, http.StatusOK, map[string]moduleItem{"item": updated})
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for index, item := range a.modules[module] {
		if item.ID != id {
			continue
		}
		updated := item
		updated.Name = input.Name
		updated.Detail = input.Detail
		updated.Status = input.Status
		if input.Map != nil {
			updated.Map = input.Map
		}
		updated.UpdatedAt = time.Now().UTC()
		a.modules[module][index] = updated
		if err := a.saveModulesLocked(); err != nil {
			a.modules[module][index] = item
			writeError(w, http.StatusInternalServerError, "No se pudo actualizar el registro")
			return
		}
		writeJSON(w, http.StatusOK, map[string]moduleItem{"item": updated})
		return
	}
	writeError(w, http.StatusNotFound, "Registro no encontrado")
}

func (a *app) deleteModuleItem(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.currentUser(r); !ok {
		writeError(w, http.StatusUnauthorized, "Sesión no iniciada")
		return
	}
	module, ok := moduleName(r)
	if !ok {
		writeError(w, http.StatusNotFound, "Módulo no encontrado")
		return
	}
	id, ok := moduleItemID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "Identificador de registro inválido")
		return
	}
	if a.dataStore != nil {
		err := a.dataStore.DeleteModuleItem(r.Context(), module, id)
		if err == repository.ErrNotFound {
			writeError(w, http.StatusNotFound, "Registro no encontrado")
			return
		}
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "No se pudo eliminar el registro")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for index, item := range a.modules[module] {
		if item.ID != id {
			continue
		}
		original := a.modules[module]
		updated := make([]moduleItem, 0, len(original)-1)
		updated = append(updated, original[:index]...)
		updated = append(updated, original[index+1:]...)
		a.modules[module] = updated
		if err := a.saveModulesLocked(); err != nil {
			a.modules[module] = original
			writeError(w, http.StatusInternalServerError, "No se pudo eliminar el registro")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeError(w, http.StatusNotFound, "Registro no encontrado")
}

func (a *app) currentUser(r *http.Request) (user, bool) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return user{}, false
	}
	if a.sessionStore != nil {
		account, ok, err := a.sessionStore.Get(r.Context(), cookie.Value)
		return account, err == nil && ok
	}
	a.mu.RLock()
	account, ok := a.sessions[cookie.Value]
	a.mu.RUnlock()
	return account, ok
}

func requestID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id > 0
}

func moduleName(r *http.Request) (string, bool) {
	module := r.PathValue("module")
	_, ok := moduleStatuses[module]
	return module, ok
}

func moduleItemID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id > 0
}

func validModuleInput(w http.ResponseWriter, r *http.Request, module string) (moduleItemInput, bool) {
	var input moduleItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "Datos de registro inválidos")
		return moduleItemInput{}, false
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Detail = strings.TrimSpace(input.Detail)
	input.Status = strings.TrimSpace(input.Status)
	if len(input.Name) < 3 || len(input.Name) > 120 || len(input.Detail) < 5 || len(input.Detail) > 600 || !moduleStatuses[module][input.Status] {
		writeError(w, http.StatusBadRequest, "Completa nombre, detalle y un estado válido")
		return moduleItemInput{}, false
	}
	if input.Map != nil && (module != "network" || !validMapMetadata(input.Map)) {
		writeError(w, http.StatusBadRequest, "La ubicación del mapa no es válida")
		return moduleItemInput{}, false
	}
	return input, true
}

func validMapMetadata(metadata *domain.MapMetadata) bool {
	metadata.Type = strings.TrimSpace(metadata.Type)
	metadata.Zone = strings.TrimSpace(metadata.Zone)
	validTypes := map[string]bool{"OLT": true, "ONT": true, "CABLE": true, "CODO": true, "DISTRITO": true}
	return validTypes[metadata.Type] && len(metadata.Zone) >= 2 && len(metadata.Zone) <= 80 && metadata.X >= 0 && metadata.X <= 100 && metadata.Y >= 0 && metadata.Y <= 100
}

func validStatus(status string) bool {
	switch status {
	case "Pendiente", "En proceso", "Resuelta", "Cancelada":
		return true
	default:
		return false
	}
}

func validCrew(crew string) bool {
	switch crew {
	case "Por asignar", "Cuadrilla Centro", "Cuadrilla Norte", "Cuadrilla Sur", "Soporte NOC":
		return true
	default:
		return false
	}
}

func (a *app) load() error {
	if _, err := repository.ReadJSON(a.requestsPath(), &a.requests); err != nil {
		return err
	}
	return a.loadModules()
}

func (a *app) saveLocked() error {
	return repository.WriteJSON(a.requestsPath(), a.requests)
}

func (a *app) loadModules() error {
	var modules map[string][]moduleItem
	exists, err := repository.ReadJSON(a.modulesPath(), &modules)
	if !exists && err == nil {
		a.modules = defaultModules()
		return a.saveModulesLocked()
	}
	if err != nil {
		return err
	}
	if modules == nil {
		a.modules = emptyModules()
		return nil
	}
	a.modules = modules
	for module := range moduleStatuses {
		if a.modules[module] == nil {
			a.modules[module] = []moduleItem{}
		}
	}
	return nil
}

func (a *app) saveModulesLocked() error {
	return repository.WriteJSON(a.modulesPath(), a.modules)
}

func (a *app) requestsPath() string {
	if a.requestsFile != "" {
		return a.requestsFile
	}
	return "data/solicitudes.json"
}

func (a *app) modulesPath() string {
	if a.moduleDataFile != "" {
		return a.moduleDataFile
	}
	return "data/modulos.json"
}

func emptyModules() map[string][]moduleItem {
	return map[string][]moduleItem{
		"network": {}, "customers": {}, "coverage": {}, "inventory": {}, "commercial": {},
	}
}

func defaultModules() map[string][]moduleItem {
	createdAt := time.Date(2026, time.November, 13, 0, 0, 0, 0, time.UTC)
	return map[string][]moduleItem{
		"network": {
			{ID: 1, Name: "OLT-QRO-CENTRO-01", Detail: "Huawei MA5800-X7 · 612 ONT en línea · óptica estable en PON 0/1", Status: "Operativa", Map: &domain.MapMetadata{Type: "OLT", Zone: "Centro", X: 20, Y: 30}, CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 2, Name: "OLT-QRO-NORTE-01", Detail: "ZTE C320 · 428 ONT en línea · potencia PON 0/2 en revisión", Status: "Atención", Map: &domain.MapMetadata{Type: "OLT", Zone: "Norte", X: 77, Y: 27}, CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 3, Name: "OLT-QRO-SUR-01", Detail: "Huawei MA5800-X2 · 376 ONT en línea · mantenimiento 23:00 h", Status: "En revisión", Map: &domain.MapMetadata{Type: "OLT", Zone: "Sur", X: 68, Y: 74}, CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 4, Name: "OLT-QRO-JURIQUILLA-01", Detail: "ZTE C320 · 295 ONT en línea · capacidad disponible 41%", Status: "Operativa", Map: &domain.MapMetadata{Type: "OLT", Zone: "Juriquilla", X: 37, Y: 18}, CreatedAt: createdAt, UpdatedAt: createdAt},
		},
		"customers": {
			{ID: 1, Name: "May Torres", Detail: "may.torres@teracom.mx|Administrador|TERACOM|442 214 6101", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 2, Name: "Roberto Méndez", Detail: "roberto.mendez@teracom.mx|Líder|TERACOM|442 214 6102", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 3, Name: "Fernanda Ruiz", Detail: "fernanda.ruiz@teracom.mx|Técnico|TERACOM|442 214 6103", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 4, Name: "Daniel Ortega", Detail: "daniel.ortega@teracom.mx|Técnico|TERACOM|442 214 6104", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 5, Name: "Paola Sosa", Detail: "paola.sosa@teracom.mx|Capturista|TERACOM|442 214 6105", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 6, Name: "Iván Castro", Detail: "ivan.castro@teracom.mx|Sub-técnico|TERACOM|442 214 6106", Status: "Inactivo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 7, Name: "Andrea Molina", Detail: "andrea.molina@teracom.mx|Cliente|TERACOM|442 214 6107", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
		},
		"coverage": {
			{ID: 1, Name: "Fraccionamiento El Refugio", Detail: "Factibilidad para 186 viviendas · levantamiento de postes en curso", Status: "En análisis", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 2, Name: "Parque Industrial Benito Juárez", Detail: "22 empresas · troncal de 2.4 km autorizada", Status: "Aprobada", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 3, Name: "Centro Histórico", Detail: "Validación de canalización y permisos municipales", Status: "Pendiente", CreatedAt: createdAt, UpdatedAt: createdAt},
		},
		"inventory": {
			{ID: 1, Name: "ONT GPON Wi‑Fi 6", Detail: "84 disponibles · punto de reorden: 40 unidades", Status: "Disponible", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 2, Name: "Caja NAP exterior de 8 puertos", Detail: "26 disponibles · reserva para expansión El Refugio", Status: "Disponible", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 3, Name: "Splitter PLC 1:8 con conectores SC/APC", Detail: "18 disponibles · punto de reorden: 25 unidades", Status: "Reabastecer", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 4, Name: "Cable drop G.657A2 de 2 fibras", Detail: "2,400 m en tránsito · recepción estimada el viernes", Status: "En tránsito", CreatedAt: createdAt, UpdatedAt: createdAt},
		},
		"commercial": {
			{ID: 1, Name: "TELMEX", Detail: "442 000 1201|operaciones.demo@teracom.mx|Cuenta de demostración · Centro|20.5924,-100.3886|Enlace empresarial 500 Mbps", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 2, Name: "Totalplay", Detail: "442 000 1202|cuenta.demo@teracom.mx|Cuenta de demostración · Norte|20.6328,-100.4021|Enlace empresarial 300 Mbps", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 3, Name: "Megacable", Detail: "442 000 1203|soporte.demo@teracom.mx|Cuenta de demostración · Sur|20.5612,-100.3778|Enlace empresarial 500 Mbps", Status: "Pendiente de pago", CreatedAt: createdAt, UpdatedAt: createdAt},
			{ID: 4, Name: "izzi", Detail: "442 000 1204|administracion.demo@teracom.mx|Cuenta de demostración · Juriquilla|20.6254,-100.4416|Enlace empresarial 1 Gbps", Status: "Activo", CreatedAt: createdAt, UpdatedAt: createdAt},
		},
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func randomID() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
