// Package repository centraliza el acceso a los archivos que persisten la operación.
package repository

import (
	"encoding/json"

	"ftth/backend/database"
)

// ReadJSON carga un documento JSON. exists permite distinguir un archivo vacío
// de uno que todavía no se ha creado.
func ReadJSON[T any](path string, target *T) (exists bool, err error) {
	contents, exists, err := database.Read(path)
	if err != nil || !exists || len(contents) == 0 {
		return exists, err
	}
	return true, json.Unmarshal(contents, target)
}

// WriteJSON serializa y guarda un documento JSON con formato legible.
func WriteJSON[T any](path string, value T) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return database.Write(path, contents)
}
