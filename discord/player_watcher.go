package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/commands"
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

// InitPlayerWatcher initializes the player watcher
func InitPlayerWatcher(s *discordgo.Session) {
	fmt.Println("Player Watcher is now running.")
	if support.Config.PlayerWatcherTargetChannelID == "" {
		fmt.Println("  Warning: PlayerWatcherTargetChannelID not configured")
		return
	}
}

// ProcessPlayerJoin handles a player joining the server (called from log.go)
// This is called for normal (non-ghost) players
func ProcessPlayerJoin(playerName string) {
	fmt.Printf("Player Watcher: %s joined\n", playerName)

	ActivePlayers.Lock()
	ActivePlayers.players[playerName] = &PlayerInfo{JoinTime: time.Now()}
	ActivePlayers.Unlock()

	// Send join message via PlayerWatcher channel
	SendPlayerWatcherJoin(playerName)
}

// ProcessPlayerLeave handles a player leaving the server (called from log.go)
// This is called for normal (non-ghost) players
func ProcessPlayerLeave(playerName string) {
	fmt.Printf("Player Watcher: %s left\n", playerName)

	var playTime string
	ActivePlayers.Lock()
	if info, exists := ActivePlayers.players[playerName]; exists {
		playTime = formatDuration(time.Since(info.JoinTime))
		delete(ActivePlayers.players, playerName)
	}
	ActivePlayers.Unlock()

	// Send leave message via PlayerWatcher channel
	SendPlayerWatcherLeave(playerName, playTime)
}

// SendPlayerWatcherJoin sends a join message to the PlayerWatcher channel
func SendPlayerWatcherJoin(playerName string) {
	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}

	message := support.Config.Messages.PlayerWatcherJoin
	if message == "" {
		message = "**Join:**\n{username} hat sich eingeloggt auf dem Server"
	}
	message = strings.ReplaceAll(message, "{username}", playerName)

	support.SendTo(Session, message, support.Config.PlayerWatcherTargetChannelID)
}

// SendPlayerWatcherLeave sends a leave message to the PlayerWatcher channel
func SendPlayerWatcherLeave(playerName string, playTime string) {
	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}

	var message string
	if playTime != "" {
		message = support.Config.Messages.PlayerWatcherLeaveTime
		if message == "" {
			message = "**Leave:**\n{username} hat sich ausgeloggt aus dem Server (Spielzeit: {playtime})"
		}
		message = strings.ReplaceAll(message, "{playtime}", playTime)
	} else {
		message = support.Config.Messages.PlayerWatcherLeave
		if message == "" {
			message = "**Leave:**\n{username} hat sich ausgeloggt aus dem Server"
		}
	}
	message = strings.ReplaceAll(message, "{username}", playerName)

	support.SendTo(Session, message, support.Config.PlayerWatcherTargetChannelID)
}

// ProcessServerInGame handles the server entering InGame state
func ProcessServerInGame() {
	fmt.Println("Player Watcher: Server is now InGame")

	if support.Config.PlayerWatcherTargetChannelID == "" {
		return
	}
	support.SendTo(Session, support.Config.Messages.ServerReady, support.Config.PlayerWatcherTargetChannelID)
}

// HandlePlayerCommand handles the player command (prefix + "player")
func HandlePlayerCommand(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	content := strings.TrimSpace(m.Content)
	prefix := support.Config.Prefix

	// Check for player command
	if !strings.HasPrefix(strings.ToLower(content), strings.ToLower(prefix+"player")) {
		return false
	}

	// Ignore bot messages for commands
	if m.Author.Bot {
		return false
	}

	// Check if admin version requested (player_admin)
	isAdminRequest := strings.HasPrefix(strings.ToLower(content), strings.ToLower(prefix+"player_admin"))
	showHiddenPlayers := false

	if isAdminRequest {
		// Check if user is admin
		if commands.CheckAdmin(m.Author.ID) {
			showHiddenPlayers = true
		}
	}

	now := time.Now()

	var players map[string]time.Time
	if showHiddenPlayers {
		// Show ALL players including hidden ones
		players = GetActivePlayers()
	} else {
		// Show only visible players (exclude ghost mode)
		players = GetVisiblePlayers()
	}

	playerCount := len(players)

	if playerCount == 0 {
		emptyMsg := support.Config.Messages.PlayerListEmpty
		if emptyMsg == "" {
			emptyMsg = "Es sind derzeit keine Spieler online."
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, emptyMsg)
		return true
	}

	var lines []string
	entryTemplate := support.Config.Messages.PlayerListEntry
	if entryTemplate == "" {
		entryTemplate = "- {username}: online seit {playtime}"
	}

	for name, joinTime := range players {
		dur := formatDuration(now.Sub(joinTime))
		entry := strings.ReplaceAll(entryTemplate, "{username}", name)
		entry = strings.ReplaceAll(entry, "{playtime}", dur)

		// Mark hidden players for admin view
		if showHiddenPlayers && IsPlayerHidden(name) {
			entry += " 👻"
		}

		lines = append(lines, entry)
	}

	headerTemplate := support.Config.Messages.PlayerListHeader
	if headerTemplate == "" {
		headerTemplate = "**Aktive Spieler:** {count}"
	}
	header := strings.ReplaceAll(headerTemplate, "{count}", fmt.Sprintf("%d", playerCount))

	if showHiddenPlayers {
		visibleCount := len(GetVisiblePlayers())
		hiddenCount := playerCount - visibleCount
		if hiddenCount > 0 {
			header += fmt.Sprintf(" (davon %d 👻 versteckt)", hiddenCount)
		}
	}

	reply := header + "\n" + strings.Join(lines, "\n")
	_, _ = s.ChannelMessageSend(m.ChannelID, reply)
	return true
}

// GetActivePlayers returns a copy of the active players map (ALL players)
func GetActivePlayers() map[string]time.Time {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()

	result := make(map[string]time.Time)
	for name, info := range ActivePlayers.players {
		result[name] = info.JoinTime
	}
	return result
}

// GetActivePlayerCount returns the number of active players (ALL players)
func GetActivePlayerCount() int {
	ActivePlayers.RLock()
	defer ActivePlayers.RUnlock()
	return len(ActivePlayers.players)
}
