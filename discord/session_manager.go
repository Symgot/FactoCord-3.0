package discord

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
)

// SessionManager verwaltet Bearbeitungssessions für Benutzer
type SessionManager struct {
	sessions map[string]*models.Session
	mu       sync.RWMutex
	basePath string
}

var globalSessionManager *SessionManager

// InitSessionManager initialisiert den globalen SessionManager
func InitSessionManager(basePath string) {
	globalSessionManager = &SessionManager{
		sessions: make(map[string]*models.Session),
		basePath: basePath,
	}

	// Erstelle Verzeichnisse
	os.MkdirAll(filepath.Join(basePath, "temp_settings"), 0755)
}

// GetSessionManager gibt den globalen SessionManager zurück
func GetSessionManager() *SessionManager {
	return globalSessionManager
}

// CreateSession erstellt eine neue Session für einen Benutzer
func (sm *SessionManager) CreateSession(userID, modName string) *models.Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Lösche vorhandene Session für diesen Benutzer
	if existingSession, exists := sm.sessions[userID]; exists {
		if existingSession.ModName != modName {
			// Anderer Mod - lösche alte Session
			sm.deleteTempFile(userID)
		}
	}

	session := models.NewSession(userID, modName)
	sm.sessions[userID] = session

	return session
}

// GetSession gibt die Session eines Benutzers zurück
func (sm *SessionManager) GetSession(userID string) *models.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[userID]
}

// UpdateSetting aktualisiert ein Setting in der Session
func (sm *SessionManager) UpdateSetting(userID, key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return &SessionError{Message: "keine aktive Session gefunden"}
	}

	session.UpdateChange(key, value)
	return sm.saveTempSettings(userID, session)
}

// RemoveSetting entfernt ein Setting aus der Session
func (sm *SessionManager) RemoveSetting(userID, key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return &SessionError{Message: "keine aktive Session gefunden"}
	}

	session.RemoveChange(key)
	return sm.saveTempSettings(userID, session)
}

// saveTempSettings speichert die Session temporär
func (sm *SessionManager) saveTempSettings(userID string, session *models.Session) error {
	tempDir := filepath.Join(sm.basePath, "temp_settings")
	tempFile := filepath.Join(tempDir, userID+".json")

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tempFile, data, 0644)
}

// CancelSession verwirft eine Session und alle Änderungen
func (sm *SessionManager) CancelSession(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[userID]; !exists {
		return &SessionError{Message: "keine aktive Session gefunden"}
	}

	delete(sm.sessions, userID)
	return sm.deleteTempFile(userID)
}

// deleteTempFile löscht die temporäre Session-Datei
func (sm *SessionManager) deleteTempFile(userID string) error {
	tempFile := filepath.Join(sm.basePath, "temp_settings", userID+".json")
	err := os.Remove(tempFile)
	if os.IsNotExist(err) {
		return nil // Datei existiert nicht, das ist OK
	}
	return err
}

// FinalizeSession schließt eine Session ab und gibt die Daten zurück
func (sm *SessionManager) FinalizeSession(userID string) (*models.Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return nil, &SessionError{Message: "keine aktive Session gefunden"}
	}

	// Kopie der Session erstellen
	sessionCopy := *session

	// Session löschen
	delete(sm.sessions, userID)
	sm.deleteTempFile(userID)

	return &sessionCopy, nil
}

// GetAllSessions gibt alle aktiven Sessions zurück (für Debug/Admin)
func (sm *SessionManager) GetAllSessions() map[string]*models.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*models.Session)
	for k, v := range sm.sessions {
		result[k] = v
	}
	return result
}

// HasSession prüft ob ein Benutzer eine aktive Session hat
func (sm *SessionManager) HasSession(userID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.sessions[userID]
	return exists
}

// LoadTempSession lädt eine Session aus der temporären Datei
func (sm *SessionManager) LoadTempSession(userID string) (*models.Session, error) {
	tempFile := filepath.Join(sm.basePath, "temp_settings", userID+".json")

	data, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, err
	}

	var session models.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	// Session in Memory laden
	sm.mu.Lock()
	sm.sessions[userID] = &session
	sm.mu.Unlock()

	return &session, nil
}

// SessionError ist ein Fehlertyp für Session-Operationen
type SessionError struct {
	Message string
}

func (e *SessionError) Error() string {
	return e.Message
}
