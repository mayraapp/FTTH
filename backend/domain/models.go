package domain

import "time"

type User struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Company string `json:"company"`
}

type LoginInput struct {
	Company  string `json:"company"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestInput struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Crew        string `json:"crew"`
	Description string `json:"description"`
}

type UpdateRequestInput struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Crew        string `json:"crew"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Request struct {
	ID          int       `json:"id"`
	Folio       string    `json:"folio"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Crew        string    `json:"crew"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type ModuleItemInput struct {
	Name   string       `json:"name"`
	Detail string       `json:"detail"`
	Status string       `json:"status"`
	Map    *MapMetadata `json:"map,omitempty"`
}

type ModuleItem struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Detail    string       `json:"detail"`
	Status    string       `json:"status"`
	Map       *MapMetadata `json:"map,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// MapMetadata conserva los datos mínimos para ubicar un elemento FTTH en GIS.
// Se mantiene opcional para no alterar los registros genéricos ya existentes.
type MapMetadata struct {
	Type string  `json:"type"`
	Zone string  `json:"zone"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

var ModuleStatuses = map[string]map[string]bool{
	"network":    {"Operativa": true, "Atención": true, "En revisión": true},
	"customers":  {"Activo": true, "Inactivo": true},
	"coverage":   {"En análisis": true, "Aprobada": true, "Pendiente": true},
	"inventory":  {"Disponible": true, "Reabastecer": true, "En tránsito": true},
	"commercial": {"Activo": true, "Pendiente de pago": true, "Suspendido": true},
}
