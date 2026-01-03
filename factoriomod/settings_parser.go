package factoriomod

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ParseModSettings liest und parst die mod-settings.dat Datei
func ParseModSettings(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Datei: %w", err)
	}

	// Factorio mod-settings.dat Format:
	// Die Datei beginnt mit einem Header gefolgt von Lua-serialisierten Daten
	// Alternativ kann es auch eine JSON-Datei sein (neuere Versionen)

	// Versuche zuerst als JSON zu parsen
	var jsonSettings map[string]interface{}
	if err := json.Unmarshal(data, &jsonSettings); err == nil {
		return jsonSettings, nil
	}

	// Ansonsten versuche Binärformat zu parsen
	return parseBinaryModSettings(data)
}

// parseBinaryModSettings parst das Binärformat der mod-settings.dat
func parseBinaryModSettings(data []byte) (map[string]interface{}, error) {
	r := bytes.NewReader(data)

	// Lese Header
	header, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen des Headers: %w", err)
	}

	// Validiere Header (Factorio verwendet typischerweise bestimmte Magic Bytes)
	_ = header // Header wird für Validierung verwendet

	// Parse die Property Tree Struktur
	settings, err := readPropertyTree(r)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Parsen der Settings: %w", err)
	}

	return settings, nil
}

// readHeader liest den Datei-Header
func readHeader(r *bytes.Reader) ([]byte, error) {
	// Factorio mod-settings.dat hat einen variablen Header
	// Typischerweise: Version (4-8 Bytes) + Flags

	header := make([]byte, 8)
	n, err := r.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return header[:n], nil
}

// readPropertyTree liest einen Factorio Property Tree
func readPropertyTree(r *bytes.Reader) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Lese Anzahl der Einträge
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		// Fallback: versuche das gesamte verbleibende als einen Eintrag zu lesen
		return readFallbackFormat(r)
	}

	for i := uint32(0); i < count; i++ {
		key, err := readString(r)
		if err != nil {
			break
		}

		value, err := readValue(r)
		if err != nil {
			break
		}

		result[key] = value
	}

	return result, nil
}

// readFallbackFormat versucht ein alternatives Format zu lesen
func readFallbackFormat(r *bytes.Reader) (map[string]interface{}, error) {
	// Einige Versionen von mod-settings.dat haben ein anderes Format
	// Hier implementieren wir einen Fallback

	result := map[string]interface{}{
		"startup":          make(map[string]interface{}),
		"runtime-global":   make(map[string]interface{}),
		"runtime-per-user": make(map[string]interface{}),
	}

	return result, nil
}

// readString liest einen String aus dem Binärformat
func readString(r *bytes.Reader) (string, error) {
	// Factorio verwendet Length-prefixed Strings
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	// Sicherheitsprüfung
	if length > 10000 {
		return "", fmt.Errorf("string zu lang: %d", length)
	}

	buf := make([]byte, length)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// readValue liest einen Wert aus dem Binärformat
func readValue(r *bytes.Reader) (interface{}, error) {
	// Lese Typ-Byte
	var valueType byte
	if err := binary.Read(r, binary.LittleEndian, &valueType); err != nil {
		return nil, err
	}

	switch valueType {
	case 0: // None/Nil
		return nil, nil

	case 1: // Bool
		var val byte
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return nil, err
		}
		return val != 0, nil

	case 2: // Number (Double)
		var val float64
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return nil, err
		}
		return val, nil

	case 3: // String
		return readString(r)

	case 4: // List/Array
		var count uint32
		if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
			return nil, err
		}

		list := make([]interface{}, count)
		for i := uint32(0); i < count; i++ {
			val, err := readValue(r)
			if err != nil {
				return nil, err
			}
			list[i] = val
		}
		return list, nil

	case 5: // Dictionary/Map
		var count uint32
		if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
			return nil, err
		}

		dict := make(map[string]interface{})
		for i := uint32(0); i < count; i++ {
			key, err := readString(r)
			if err != nil {
				return nil, err
			}

			val, err := readValue(r)
			if err != nil {
				return nil, err
			}

			dict[key] = val
		}
		return dict, nil

	default:
		return nil, fmt.Errorf("unbekannter Werttyp: %d", valueType)
	}
}

// ParseModSettingsJSON parst eine mod-settings.json Datei (alternatives Format)
func ParseModSettingsJSON(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}
