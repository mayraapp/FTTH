package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateRequestPersistsLocally(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}}}
	body := bytes.NewBufferString(`{"type":"Mantenimiento","priority":"Alta","crew":"Cuadrilla Centro","description":"Se requiere revisar el enlace principal de fibra."}`)
	req := httptest.NewRequest(http.MethodPost, "/api/solicitudes", body)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.createRequest(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("estado esperado 201, recibido %d: %s", response.Code, response.Body.String())
	}
	if _, err := os.Stat(filepath.Join("data", "solicitudes.json")); err != nil {
		t.Fatalf("la solicitud no se persistió: %v", err)
	}

	var payload struct {
		Request request `json:"request"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Request.Folio != "FTTH-0001" || payload.Request.Status != "Pendiente" || payload.Request.Crew != "Cuadrilla Centro" {
		t.Fatalf("solicitud inesperada: %+v", payload.Request)
	}
}

func TestCreateRequestAppliesQuickIntakeDefaults(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}}}
	req := httptest.NewRequest(http.MethodPost, "/api/solicitudes", bytes.NewBufferString(`{"type":"Instalación","description":"Instalar servicio de fibra en domicilio del cliente."}`))
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.createRequest(response, req)
	if response.Code != http.StatusCreated {
		t.Fatalf("estado esperado 201, recibido %d: %s", response.Code, response.Body.String())
	}
	if got := a.requests[0]; got.Priority != "Media" || got.Crew != "Por asignar" || got.Status != "Pendiente" {
		t.Fatalf("valores iniciales inesperados: %+v", got)
	}
}

func TestCreateRequestRequiresSession(t *testing.T) {
	a := &app{sessions: make(map[string]user)}
	req := httptest.NewRequest(http.MethodPost, "/api/solicitudes", bytes.NewBufferString(`{}`))
	response := httptest.NewRecorder()

	a.createRequest(response, req)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("estado esperado 401, recibido %d", response.Code)
	}
}

func TestLogoutInvalidatesSessionAndExpiresCookie(t *testing.T) {
	a := &app{sessions: map[string]user{"test-session": {Name: "May Torres", Role: "Administrador", Company: "TERACOM"}}}
	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	logoutRequest.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	logoutResponse := httptest.NewRecorder()

	a.logout(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("estado esperado 204, recibido %d", logoutResponse.Code)
	}
	if _, exists := a.sessions["test-session"]; exists {
		t.Fatal("la sesión debe eliminarse del servidor")
	}

	var expiredCookie *http.Cookie
	for _, cookie := range logoutResponse.Result().Cookies() {
		if cookie.Name == sessionName {
			expiredCookie = cookie
			break
		}
	}
	if expiredCookie == nil || expiredCookie.MaxAge >= 0 || expiredCookie.Path != "/" {
		t.Fatalf("cookie de expiración inválida: %+v", expiredCookie)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meRequest.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	meResponse := httptest.NewRecorder()
	a.me(meResponse, meRequest)
	if meResponse.Code != http.StatusUnauthorized {
		t.Fatalf("la sesión cerrada no debe recuperar acceso; recibido %d", meResponse.Code)
	}
}

func TestListRequestsReturnsEmptyArray(t *testing.T) {
	a := &app{sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}}}
	req := httptest.NewRequest(http.MethodGet, "/api/solicitudes", nil)
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.listRequests(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("estado esperado 200, recibido %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Requests []request `json:"requests"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Requests == nil || len(payload.Requests) != 0 {
		t.Fatalf("se esperaba un arreglo vacío, recibido: %s", response.Body.String())
	}
}

func TestUpdateRequestPersistsChanges(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{
		requests: []request{{ID: 1, Folio: "FTTH-0001", Type: "Mantenimiento", Priority: "Alta", Description: "Revisar enlace principal de fibra.", Status: "Pendiente"}},
		sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}},
	}
	req := httptest.NewRequest(http.MethodPut, "/api/solicitudes/1", bytes.NewBufferString(`{"type":"Incidencia de red","priority":"Media","crew":"Cuadrilla Norte","description":"La señal óptica presenta variaciones en el sector norte.","status":"En proceso"}`))
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.updateRequest(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("estado esperado 200, recibido %d: %s", response.Code, response.Body.String())
	}
	if got := a.requests[0]; got.Type != "Incidencia de red" || got.Priority != "Media" || got.Crew != "Cuadrilla Norte" || got.Status != "En proceso" {
		t.Fatalf("solicitud no actualizada: %+v", got)
	}
	contents, err := os.ReadFile(filepath.Join("data", "solicitudes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(contents, []byte(`"status": "En proceso"`)) {
		t.Fatalf("actualización no persistida: %s", contents)
	}
}

func TestDeleteRequestPersistsChanges(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{
		requests: []request{{ID: 1, Folio: "FTTH-0001", Type: "Mantenimiento", Priority: "Alta", Description: "Revisar enlace principal de fibra.", Status: "Pendiente"}},
		sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/solicitudes/1", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.deleteRequest(response, req)
	if response.Code != http.StatusNoContent {
		t.Fatalf("estado esperado 204, recibido %d: %s", response.Code, response.Body.String())
	}
	if len(a.requests) != 0 {
		t.Fatalf("la solicitud no se eliminó: %+v", a.requests)
	}
	contents, err := os.ReadFile(filepath.Join("data", "solicitudes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "[]" {
		t.Fatalf("eliminación no persistida: %s", contents)
	}
}

func TestModuleItemCRUDPersistsChanges(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{
		modules:  emptyModules(),
		sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}},
	}
	create := httptest.NewRequest(http.MethodPost, "/api/modulos/commercial", bytes.NewBufferString(`{"name":"Contrato PRUEBA-01","detail":"Cuenta de prueba para validar la persistencia local.","status":"Activo"}`))
	create.SetPathValue("module", "commercial")
	create.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	createdResponse := httptest.NewRecorder()

	a.createModuleItem(createdResponse, create)
	if createdResponse.Code != http.StatusCreated {
		t.Fatalf("estado esperado 201, recibido %d: %s", createdResponse.Code, createdResponse.Body.String())
	}
	if len(a.modules["commercial"]) != 1 || a.modules["commercial"][0].Status != "Activo" {
		t.Fatalf("registro de módulo inesperado: %+v", a.modules["commercial"])
	}
	if _, err := os.Stat(filepath.Join("data", "modulos.json")); err != nil {
		t.Fatalf("el registro no se persistió: %v", err)
	}

	update := httptest.NewRequest(http.MethodPut, "/api/modulos/commercial/1", bytes.NewBufferString(`{"name":"Contrato PRUEBA-01","detail":"Cuenta revisada y preparada para la operación.","status":"Pendiente de pago"}`))
	update.SetPathValue("module", "commercial")
	update.SetPathValue("id", "1")
	update.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	updatedResponse := httptest.NewRecorder()
	a.updateModuleItem(updatedResponse, update)
	if updatedResponse.Code != http.StatusOK || a.modules["commercial"][0].Status != "Pendiente de pago" {
		t.Fatalf("registro de módulo no actualizado: %d %+v", updatedResponse.Code, a.modules["commercial"])
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/modulos/commercial/1", nil)
	deleteRequest.SetPathValue("module", "commercial")
	deleteRequest.SetPathValue("id", "1")
	deleteRequest.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	deletedResponse := httptest.NewRecorder()
	a.deleteModuleItem(deletedResponse, deleteRequest)
	if deletedResponse.Code != http.StatusNoContent || len(a.modules["commercial"]) != 0 {
		t.Fatalf("registro de módulo no eliminado: %d %+v", deletedResponse.Code, a.modules["commercial"])
	}
}

func TestModuleItemPersistsMapMetadataAndPreservesItOnEdit(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	a := &app{
		modules:  emptyModules(),
		sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}},
	}
	create := httptest.NewRequest(http.MethodPost, "/api/modulos/network", bytes.NewBufferString(`{"name":"OLT-JARDINES-01","detail":"Equipo de acceso en Jardines.","status":"Operativa","map":{"type":"OLT","zone":"Jardines","x":64,"y":38}}`))
	create.SetPathValue("module", "network")
	create.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	createdResponse := httptest.NewRecorder()
	a.createModuleItem(createdResponse, create)

	if createdResponse.Code != http.StatusCreated {
		t.Fatalf("estado esperado 201, recibido %d: %s", createdResponse.Code, createdResponse.Body.String())
	}
	created := a.modules["network"][0]
	if created.Map == nil || created.Map.Type != "OLT" || created.Map.Zone != "Jardines" || created.Map.X != 64 || created.Map.Y != 38 {
		t.Fatalf("metadatos de mapa inesperados: %+v", created.Map)
	}

	update := httptest.NewRequest(http.MethodPut, "/api/modulos/network/1", bytes.NewBufferString(`{"name":"OLT-JARDINES-01","detail":"Equipo de acceso revisado en Jardines.","status":"Atención"}`))
	update.SetPathValue("module", "network")
	update.SetPathValue("id", "1")
	update.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	updatedResponse := httptest.NewRecorder()
	a.updateModuleItem(updatedResponse, update)

	updated := a.modules["network"][0]
	if updatedResponse.Code != http.StatusOK || updated.Map == nil || updated.Map.Type != "OLT" || updated.Map.X != 64 || updated.Map.Y != 38 {
		t.Fatalf("la edición no preservó la ubicación: %d %+v", updatedResponse.Code, updated.Map)
	}
}

func TestListModuleItemsReturnsEmptyArray(t *testing.T) {
	a := &app{
		modules:  emptyModules(),
		sessions: map[string]user{"test-session": {Name: "Administradora", Role: "Administrador", Company: "TERACOM"}},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/modulos/inventory", nil)
	request.SetPathValue("module", "inventory")
	request.AddCookie(&http.Cookie{Name: sessionName, Value: "test-session"})
	response := httptest.NewRecorder()

	a.listModuleItems(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("estado esperado 200, recibido %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Items []moduleItem `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Items == nil || len(payload.Items) != 0 {
		t.Fatalf("se esperaba un arreglo vacío, recibido: %s", response.Body.String())
	}
}
