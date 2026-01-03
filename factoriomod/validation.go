package factoriomod

import (
	"fmt"
	"reflect"
)

// ValidationError repräsentiert einen Validierungsfehler
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateSetting validiert einen Setting-Wert
func ValidateSetting(modName, key string, value interface{}) error {
	mod := GlobalModManager.GetModByName(modName)
	if mod == nil {
		return &ValidationError{
			Field:   modName,
			Message: "Mod nicht gefunden",
		}
	}

	// Prüfe ob das Setting existiert
	existingValue := findSettingValue(mod, key)
	if existingValue == nil {
		return &ValidationError{
			Field:   key,
			Message: "Setting nicht gefunden",
		}
	}

	// Typ-Validierung
	if err := validateType(key, existingValue, value); err != nil {
		return err
	}

	// Wert-Validierung
	if err := validateValue(key, value); err != nil {
		return err
	}

	return nil
}

// findSettingValue findet einen Setting-Wert in einem Mod
func findSettingValue(mod interface{}, key string) interface{} {
	// Implementierung je nach Mod-Struktur
	// Dies ist eine vereinfachte Version
	return nil
}

// validateType prüft ob der neue Wert den richtigen Typ hat
func validateType(key string, existingValue, newValue interface{}) error {
	existingType := reflect.TypeOf(existingValue)
	newType := reflect.TypeOf(newValue)

	// Erlaube kompatible Typen
	if existingType == nil || newType == nil {
		return nil
	}

	// Prüfe numerische Typen
	if isNumericType(existingType) && isNumericType(newType) {
		return nil
	}

	// Prüfe exakte Übereinstimmung
	if existingType.Kind() != newType.Kind() {
		return &ValidationError{
			Field:   key,
			Message: fmt.Sprintf("ungültiger Typ: erwartet %s, bekommen %s", existingType.Kind(), newType.Kind()),
		}
	}

	return nil
}

// isNumericType prüft ob ein Typ numerisch ist
func isNumericType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// validateValue prüft ob ein Wert gültig ist
func validateValue(key string, value interface{}) error {
	// Null-Werte sind generell ungültig
	if value == nil {
		return &ValidationError{
			Field:   key,
			Message: "Wert darf nicht nil sein",
		}
	}

	// Strings dürfen nicht leer sein (optional)
	if str, ok := value.(string); ok {
		if len(str) > 10000 {
			return &ValidationError{
				Field:   key,
				Message: "String zu lang (max 10000 Zeichen)",
			}
		}
	}

	// Numerische Werte auf sinnvolle Grenzen prüfen
	if num, ok := toFloat64(value); ok {
		if num < -1e15 || num > 1e15 {
			return &ValidationError{
				Field:   key,
				Message: "numerischer Wert außerhalb des gültigen Bereichs",
			}
		}
	}

	return nil
}

// toFloat64 konvertiert einen Wert zu float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	}
	return 0, false
}

// ValidateChanges validiert mehrere Änderungen
func ValidateChanges(modName string, changes map[string]interface{}) []error {
	var errors []error

	for key, value := range changes {
		if err := ValidateSetting(modName, key, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateSettingName prüft ob ein Setting-Name gültig ist
func ValidateSettingName(name string) error {
	if name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "Setting-Name darf nicht leer sein",
		}
	}

	if len(name) > 256 {
		return &ValidationError{
			Field:   "name",
			Message: "Setting-Name zu lang (max 256 Zeichen)",
		}
	}

	return nil
}

// ValidateModName prüft ob ein Mod-Name gültig ist
func ValidateModName(name string) error {
	if name == "" {
		return &ValidationError{
			Field:   "modName",
			Message: "Mod-Name darf nicht leer sein",
		}
	}

	if len(name) > 256 {
		return &ValidationError{
			Field:   "modName",
			Message: "Mod-Name zu lang (max 256 Zeichen)",
		}
	}

	return nil
}
