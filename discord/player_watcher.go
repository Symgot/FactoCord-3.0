package discord

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// PlayerInfo stores information about an active player
type PlayerInfo struct {
	JoinTime time.Time
}

// ActivePlayers tracks currently online players
var ActivePlayers = struct {
	sync.RWMutex
	players map[string]*PlayerInfo
}{players: make(map[string]*PlayerInfo)}

// Regex to parse log lines like: "2025-12-25 15:23:32 [JOIN] HKWN1997 joined the game"
var playerLogRegexp = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \[(JOIN|LEAVE)\] (\S+) (joined|left) the game$`)

// ParsedLogLine represents a parsed player log entry
type ParsedLogLine struct {
	Timestamp time.Time
	Type      string // "JOIN" or "LEAVE"
	Name      string
}

// parsePlayerLogLine parses a log line and returns structured data
func parsePlayerLogLine(content string) *ParsedLogLine {
	content = strings.TrimSpace(content)
	matches := playerLogRegexp.FindStringSubmatch(content)
	if matches == nil {
		return nil
	}

	// Parse timestamp (format: "2025-12-25 15:23:32")
	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		timestamp = time.Now()
	}

	return &ParsedLogLine{
		Timestamp: timestamp,
		Type:      matches[2],
		Name:      matches[3],
	}
}

// formatDuration formats a duration in a human-readable format (e.g., "2h 30m 15s")
func formatDuration(d time.Duration) string {
	totalSec := int(d.Seconds())
	hours := totalSec / 3600
	minutes := (totalSec % 3600) / 60
	seconds := totalSec % 60

	var parts []string
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	parts = append(parts, fmt.Sprintf("%ds", seconds))

	return strings.Join(parts, " ")
}

// InitPlayerWatcher sends a startup message to the target channel
func InitPlayerWatcher(s *discordgo.Session) {
	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}
	support.SendTo(s, "Bot wurde gestartet und bereit!", support.Config.PlayerWatcherTargetChannelID)
}

// HandlePlayerWatcherMessage processes messages from the source channel
// Returns true if the message was handled
func HandlePlayerWatcherMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// Check if player watcher is configured
	if support.Config.PlayerWatcherSourceChannelID == "" || support.Config.PlayerWatcherTargetChannelID == "" {
		return false
	}

	// Only handle messages from source channel
	if m.ChannelID != support.Config.PlayerWatcherSourceChannelID {
		return false
	}

	// Parse the log line
	parsed := parsePlayerLogLine(m.Content)
	if parsed == nil {
		return false
	}

	// Handle JOIN/LEAVE events
	if parsed.Type == "JOIN" {
		ActivePlayers.Lock()
		ActivePlayers.players[parsed.Name] = &PlayerInfo{JoinTime: parsed.Timestamp}
		ActivePlayers.Unlock()

		message := fmt.Sprintf("Join:\n%s hat sich eingeloggt auf dem Server", parsed.Name)
		support.SendTo(s, message, support.Config.PlayerWatcherTargetChannelID)
	} else if parsed.Type == "LEAVE" {
		ActivePlayers.Lock()
		delete(ActivePlayers.players, parsed.Name)
		ActivePlayers.Unlock()

		message := fmt.Sprintf("Leave:\n%s hat sich ausgeloggt aus dem Server", parsed.Name)
		support.SendTo(s, message, support.Config.PlayerWatcherTargetChannelID)
	}

	return true
}

// HandlePlayerCommand handles the !player command
func HandlePlayerCommand(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	content := strings.TrimSpace(strings.ToLower(m.Content))
	if !strings.HasPrefix(content, "!player") {
		return false
	}

	// Ignore bot messages for commands
	if m.Author.Bot {
		return false
	}

	now := time.Now()

	ActivePlayers.RLock()
	playerCount := len(ActivePlayers.players)
	if playerCount == 0 {
		ActivePlayers.RUnlock()
		_, _ = s.ChannelMessageSend(m.ChannelID, "Es sind derzeit keine Spieler online.")
		return true
	}

	var lines []string
	for name, info := range ActivePlayers.players {
		dur := formatDuration(now.Sub(info.JoinTime))
		lines = append(lines, fmt.Sprintf("- %s: online seit %s", name, dur))
	}
	ActivePlayers.RUnlock()

	reply := fmt.Sprintf("Aktive Spieler: %d\n%s", playerCount, strings.Join(lines, "\n"))
	_, _ = s.ChannelMessageSend(m.ChannelID, reply)
	return true
}

// GetActivePlayers returns a copy of the active players map
func GetActivePlayers() map[string]time.Time {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()

	result := make(map[string]time.Time)
	for name, info := range ActivePlayers.players {
		result[name] = info.JoinTime
	}
	return result
}

// GetActivePlayerCount returns the number of active players
func GetActivePlayerCount() int {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()
	return len(ActivePlayers.players)
}
