package models

import "time"

// Session represents an active mod settings editing session for a user
type Session struct {
	UserID      string                 `json:"user_id"`
	ModName     string                 `json:"mod_name"`
	Changes     map[string]interface{} `json:"changes"`
	CreatedAt   time.Time              `json:"created_at"`
	LastUpdated time.Time              `json:"last_updated"`
}

// NewSession creates a new session with initialized fields
func NewSession(userID, modName string) *Session {
	now := time.Now()
	return &Session{
		UserID:      userID,
		ModName:     modName,
		Changes:     make(map[string]interface{}),
		CreatedAt:   now,
		LastUpdated: now,
	}
}

// HasChanges returns true if the session has pending changes
func (s *Session) HasChanges() bool {
	return len(s.Changes) > 0
}

// UpdateChange adds or updates a setting change
func (s *Session) UpdateChange(key string, value interface{}) {
	s.Changes[key] = value
	s.LastUpdated = time.Now()
}

// RemoveChange removes a pending change
func (s *Session) RemoveChange(key string) {
	delete(s.Changes, key)
	s.LastUpdated = time.Now()
}

// ClearChanges removes all pending changes
func (s *Session) ClearChanges() {
	s.Changes = make(map[string]interface{})
	s.LastUpdated = time.Now()
}
