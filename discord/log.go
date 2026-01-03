package discord

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

var chatRegexp = regexp.MustCompile(`^\d{4}[-/]\d\d[-/]\d\d \d\d:\d\d:\d\d `)
var factorioLogRegexp = regexp.MustCompile(`^\d+\.\d{3} `)
var gameidRegexp = regexp.MustCompile("Matching server game `(\\d+)` has been created")

var forwardMessages = []*regexp.Regexp{
	regexp.MustCompile("^Player .+ doesn't exist."),
	regexp.MustCompile("^.+ wasn't banned."),
}

var consoleChannel chan string = nil

// ProcessFactorioLogLine pipes in-game chat to Discord.
func ProcessFactorioLogLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "Sendto failed (but can be probably ignored)") {
		return
	}

	// Check if any ghost mode player should have their logs suppressed
	if ShouldSuppressLog(line) {
		return // Don't forward this log at all
	}

	if support.Config.EnableConsoleChannel && support.Config.FactorioConsoleChatID != "" {
		if consoleChannel == nil {
			consoleChannel = make(chan string, 10)
			go forwardToConsoleChannel(Session, consoleChannel)
		}
		// Check again before forwarding to console channel
		if !ShouldSuppressLog(line) {
			consoleChannel <- line
		}
	}

	if chatRegexp.FindString(line) != "" {
		line = line[len("0000-00-00 00:00:00 "):]
		processFactorioChat(strings.TrimSpace(line))
	} else if factorioLogRegexp.FindString(line) != "" {
		if strings.Contains(line, "Quitting: multiplayer error.") {
			support.SendMessage(Session, support.Config.Messages.ServerFail)
		}
		if strings.Contains(line, "Opening socket for broadcast") {
			support.SendMessage(Session, support.Config.Messages.ServerStart)
		}
		if strings.Contains(line, "Saving finished") {
			if support.MyLastMessage && strings.HasPrefix(support.LastMessage.Metadata, "save") {
				num, _ := strconv.ParseInt(support.LastMessage.Metadata[len("save"):], 10, 0)
				num += 1
				support.LastMessage.Edit(Session, support.Config.Messages.ServerSave+fmt.Sprintf(" [x%d]", num))
				support.LastMessage.Metadata = fmt.Sprintf("save%d", num)
			} else {
				message := support.SendMessage(Session, support.Config.Messages.ServerSave)
				if message != nil {
					message.Metadata = "save1"
				}
				support.Factorio.SaveRequested = false
			}
		}
		if strings.Contains(line, "Quitting multiplayer connection.") {
			support.SendMessage(Session, support.Config.Messages.ServerStop)
		}
		// Detect server entering InGame state (ServerMultiplayerManager changing to InGame)
		if strings.Contains(line, "changing state from(CreatingGame) to(InGame)") {
			ProcessServerInGame()
		}
		if match := gameidRegexp.FindStringSubmatch(line); match != nil {
			support.Factorio.GameID = match[1]
		}
	} else {
		for _, pattern := range forwardMessages {
			if pattern.FindString(line) != "" {
				support.Send(Session, line)
				return
			}
		}
	}
}

var chatStartRegexp = regexp.MustCompile(`^\[(CHAT|JOIN|LEAVE|KICK|BAN|DISCORD|DISCORD-EMBED)]`)

func sendPlayerStateMessage(line, template string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 || template == "" {
		return false
	}
	username := fields[0]
	message := strings.ReplaceAll(template, "{username}", username)
	support.SendMessage(Session, message)
	return true
}

func processFactorioChat(line string) {
	match := chatStartRegexp.FindStringSubmatch(line)
	if match == nil {
		return
	}
	messageType := match[1]
	integrationMessage := messageType == "DISCORD-EMBED" || messageType == "DISCORD"

	line = strings.TrimLeft(line[len(messageType)+2:], " ")
	if strings.HasPrefix(line, "<server>") {
		return
	}
	switch messageType {
	case "JOIN":
		// Extract player name and check ghost mode
		fields := strings.Fields(line)
		if len(fields) > 0 {
			playerName := fields[0]

			// Check if this player has pre-login ghost mode
			if OnPlayerLogin(playerName) {
				// Player has ghost mode pre-login enabled
				// Don't show join message, don't add to visible players
				fmt.Printf("Player Watcher: %s joined (GHOST MODE - hidden)\n", playerName)
				return
			}

			// Check if player is in ghost mode (shouldn't happen for JOIN, but safety check)
			if IsPlayerHidden(playerName) {
				return
			}

			// ProcessPlayerJoin handles both tracking AND sending notification via PlayerWatcher
			ProcessPlayerJoin(playerName)
		}
		// Don't send player_join message here - it's handled by PlayerWatcher
		return

	case "LEAVE":
		// Extract player name and check ghost mode
		fields := strings.Fields(line)
		if len(fields) > 0 {
			playerName := fields[0]

			// Check if player is in ghost mode - if so, don't show leave
			if IsPlayerHidden(playerName) {
				fmt.Printf("Player Watcher: %s left (GHOST MODE - hidden)\n", playerName)
				// Clean up ghost mode data
				SetPlayerHidden(playerName, false)
				return
			}

			// ProcessPlayerLeave handles both tracking AND sending notification via PlayerWatcher
			ProcessPlayerLeave(playerName)
		}
		// Don't send player_leave message here - it's handled by PlayerWatcher
		return
	case "DISCORD", "CHAT":
		// Check if the message is from a hidden player
		if ShouldSuppressLog(line) {
			return
		}
		if strings.Contains(line, "@") {
			line = AddMentions(line)
			if !support.Config.AllowPingingEveryone {
				line = strings.ReplaceAll(line, "@here", "@​here")
				line = strings.ReplaceAll(line, "@everyone", "@​everyone")
			}
		}
		if messageType == "DISCORD" && support.Config.HaveServerEssentials {
			support.Send(Session, line)
			return
		}
		if !integrationMessage {
			support.Send(Session, line)
		}
	case "DISCORD-EMBED":
		if support.Config.HaveServerEssentials {
			message := new(discordgo.MessageSend)
			err := json.Unmarshal([]byte(line), message)
			if err == nil {
				message.TTS = false
				support.SendComplex(Session, message)
			}
		}
	default:
		if !integrationMessage {
			support.Send(Session, line)
		}
	}
}

func forwardToConsoleChannel(s *discordgo.Session, lines chan string) {
	message := ""
	var timeout <-chan time.Time = nil
	for {
		select {
		case <-timeout:
			support.SendTo(s, message, support.Config.FactorioConsoleChatID)
			message = ""
			timeout = nil
		case line := <-lines:
			line = strings.ReplaceAll(line, "_", "\\_")
			line = strings.ReplaceAll(line, "*", "\\*")
			line = strings.ReplaceAll(line, ">", "\\>")
			if len(message)+len(line)+1 >= 2000 {
				support.SendTo(s, message, support.Config.FactorioConsoleChatID)
				message = ""
				timeout = nil
			}
			if timeout == nil {
				timeout = time.After(time.Second * 2)
			}
			message += "\n" + line
		}
	}
}
