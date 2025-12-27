package discord

import (
	"fmt"
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
	fmt.Println("Player Watcher is now running.")
	if support.Config.PlayerWatcherTargetChannelID == "" {
		fmt.Println("  Warning: PlayerWatcherTargetChannelID not configured")
		return
	}
	support.SendTo(s, "Bot wurde gestartet und bereit!", support.Config.PlayerWatcherTargetChannelID)
}

// ProcessPlayerJoin handles a player joining the server (called from log.go)
func ProcessPlayerJoin(playerName string) {
	fmt.Printf("Player Watcher: %s joined\n", playerName)

	ActivePlayers.Lock()
	ActivePlayers.players[playerName] = &PlayerInfo{JoinTime: time.Now()}
	ActivePlayers.Unlock()

	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}
	message := fmt.Sprintf("**Join:**\n%s hat sich eingeloggt auf dem Server", playerName)
	support.SendTo(Session, message, support.Config.PlayerWatcherTargetChannelID)
}

// ProcessPlayerLeave handles a player leaving the server (called from log.go)
func ProcessPlayerLeave(playerName string) {
	fmt.Printf("Player Watcher: %s left\n", playerName)

	var playTime string
	ActivePlayers.Lock()
	if info, exists := ActivePlayers.players[playerName]; exists {
		playTime = formatDuration(time.Since(info.JoinTime))
		delete(ActivePlayers.players, playerName)
	}
	ActivePlayers.Unlock()

	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}

	var message string
	if playTime != "" {
		message = fmt.Sprintf("**Leave:**\n%s hat sich ausgeloggt aus dem Server (Spielzeit: %s)", playerName, playTime)
	} else {
		message = fmt.Sprintf("**Leave:**\n%s hat sich ausgeloggt aus dem Server", playerName)
	}
	support.SendTo(Session, message, support.Config.PlayerWatcherTargetChannelID)
}

// ProcessServerInGame handles the server entering InGame state
func ProcessServerInGame() {
	fmt.Println("Player Watcher: Server is now InGame")

	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}
	support.SendTo(Session, "**Server Status:**\nServer ist jetzt im Spiel und bereit für Spieler!", support.Config.PlayerWatcherTargetChannelID)
}

// HandlePlayerWatcherMessage processes messages from the source channel (legacy, kept for compatibility)
// Returns true if the message was handled
func HandlePlayerWatcherMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// This function is now optional - player tracking is done via ProcessFactorioLogLine
	return false
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

	reply := fmt.Sprintf("**Aktive Spieler:** %d\n%s", playerCount, strings.Join(lines, "\n"))
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
