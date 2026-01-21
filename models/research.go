package models

import (
	"encoding/json"
	"sync"
	"time"
)

type ResearchState string

const (
	ResearchedState      ResearchState = "researched"
	CurrentState         ResearchState = "current"
	AvailableDirectState ResearchState = "available_direct" // Gelb
	AvailableAfterState  ResearchState = "available_after"  // Dunkelgelb
	UnavailableState     ResearchState = "unavailable"
)

// Research repräsentiert eine einzelne Technologie
type Research struct {
	Name          string        `json:"name"`
	Level         int           `json:"level"`
	State         ResearchState `json:"state"`
	Prerequisites []string      `json:"prerequisites"`
	Cost          uint64        `json:"cost"`
	Effects       []interface{} `json:"effects"`
	LocalizedName string        `json:"localized_name,omitempty"` // Optional für UI
}

// ResearchQueue repräsentiert die Forschungsschlange
type ResearchQueueItem struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// TechTree ist die vollständige Technologie-Struktur
type TechTree struct {
	Researched []Research           `json:"researched"`
	Current    *Research            `json:"current"`
	Queue      []ResearchQueueItem  `json:"queue"`
	AllTechs   map[string]*Research `json:"tree"`
	LastUpdate time.Time            `json:"last_update"`
	Stats      *ResearchStats       `json:"stats"`
}

// ResearchStats enthält Statistiken über den Tech-Tree
type ResearchStats struct {
	TotalTechs           int `json:"total_techs"`
	ResearchedCount      int `json:"researched_count"`
	AvailableDirectCount int `json:"available_direct_count"`
	AvailableAfterCount  int `json:"available_after_count"`
	UnavailableCount     int `json:"unavailable_count"`
	QueueLength          int `json:"queue_length"`
}

var (
	// cachedTechTree speichert den letzten abgerufenen Tech-Tree
	cachedTechTree *TechTree
	techTreeMutex  sync.RWMutex

	// lastTechTreeUpdate verfolgt wann der Tree zuletzt aktualisiert wurde
	lastTechTreeUpdate time.Time
	updateMutex        sync.RWMutex
)

// UpdateTechTree aktualisiert den gecachten Tech-Tree
func UpdateTechTree(rawData []byte) error {
	var tree TechTree
	if err := json.Unmarshal(rawData, &tree); err != nil {
		return err
	}

	// Validiere und berechne Stats
	if tree.AllTechs == nil {
		tree.AllTechs = make(map[string]*Research)
	}

	// Berechne Statistiken
	tree.Stats = calculateStats(&tree)

	tree.LastUpdate = time.Now()

	// Speichere im Cache
	techTreeMutex.Lock()
	defer techTreeMutex.Unlock()
	cachedTechTree = &tree

	// Aktualisiere auch den globalen Zeitstempel
	updateMutex.Lock()
	defer updateMutex.Unlock()
	lastTechTreeUpdate = time.Now()

	return nil
}

// GetTechTree gibt den gecachten Tech-Tree zurück
func GetTechTree() *TechTree {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()
	return cachedTechTree
}

// GetTechTreeJSON gibt den gecachten Tech-Tree als JSON zurück
func GetTechTreeJSON() []byte {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil {
		return nil
	}

	data, _ := json.Marshal(cachedTechTree)
	return data
}

// GetResearchByName findet eine Technologie nach Name
func GetResearchByName(name string) *Research {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil || cachedTechTree.AllTechs == nil {
		return nil
	}

	return cachedTechTree.AllTechs[name]
}

// GetResearchesByState gibt alle Technologien eines bestimmten States zurück
func GetResearchesByState(state ResearchState) []*Research {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil || cachedTechTree.AllTechs == nil {
		return []*Research{}
	}

	var result []*Research
	for _, tech := range cachedTechTree.AllTechs {
		if tech.State == state {
			result = append(result, tech)
		}
	}
	return result
}

// GetCurrentResearch gibt die aktuelle Forschung zurück
func GetCurrentResearch() *Research {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil {
		return nil
	}

	return cachedTechTree.Current
}

// GetResearchQueue gibt die aktuelle Forschungsschlange zurück
func GetResearchQueue() []ResearchQueueItem {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil {
		return []ResearchQueueItem{}
	}

	return cachedTechTree.Queue
}

// GetResearchStats gibt die Statistiken zurück
func GetResearchStats() *ResearchStats {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()

	if cachedTechTree == nil || cachedTechTree.Stats == nil {
		return &ResearchStats{}
	}

	return cachedTechTree.Stats
}

// HasTechTreeData prüft ob Daten vorhanden sind
func HasTechTreeData() bool {
	techTreeMutex.RLock()
	defer techTreeMutex.RUnlock()
	return cachedTechTree != nil && len(cachedTechTree.AllTechs) > 0
}

// GetTechTreeAge gibt das Alter des gecachten Trees in Sekunden zurück
func GetTechTreeAge() int64 {
	updateMutex.RLock()
	defer updateMutex.RUnlock()

	if lastTechTreeUpdate.IsZero() {
		return -1
	}

	return int64(time.Since(lastTechTreeUpdate).Seconds())
}

// calculateStats berechnet Statistiken für den Tech-Tree
func calculateStats(tree *TechTree) *ResearchStats {
	stats := &ResearchStats{
		TotalTechs:  len(tree.AllTechs),
		QueueLength: len(tree.Queue),
	}

	for _, tech := range tree.AllTechs {
		switch tech.State {
		case ResearchedState:
			stats.ResearchedCount++
		case AvailableDirectState:
			stats.AvailableDirectCount++
		case AvailableAfterState:
			stats.AvailableAfterCount++
		case UnavailableState:
			stats.UnavailableCount++
		}
	}

	return stats
}

// ClearTechTree löscht den gecachten Tech-Tree
func ClearTechTree() {
	techTreeMutex.Lock()
	defer techTreeMutex.Unlock()
	cachedTechTree = nil

	updateMutex.Lock()
	defer updateMutex.Unlock()
	lastTechTreeUpdate = time.Time{}
}
