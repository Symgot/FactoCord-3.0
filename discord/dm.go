package discord

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/commands"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// DMSession holds information about a DM session with an admin
type DMSession struct {
	UserID   string
	Username string
}

// HandleDMMessage processes messages sent in DMs
func HandleDMMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	// Check if DM chat is enabled
	if !support.Config.EnableDMChat {
		return false
	}

	// Check if this is a DM channel
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return false
	}
	if channel.Type != discordgo.ChannelTypeDM {
		return false
	}

	// Don't respond to bot's own messages
	if m.Author.ID == s.State.User.ID {
		return true
	}

	// Check if user is an admin
	isAdmin := commands.CheckAdmin(m.Author.ID)

	// Handle the message
	content := strings.TrimSpace(m.Content)

	// Check for command prefix
	if strings.HasPrefix(content, support.Config.Prefix) {
		input := strings.TrimPrefix(content, support.Config.Prefix)
		handleDMCommand(s, m, input, isAdmin)
		return true
	}

	// If not a command and user is admin, show help
	if isAdmin {
		sendDMHelp(s, m.ChannelID)
	} else {
		sendDMToChannel(s, m.ChannelID, "❌ You are not authorized to use DM commands. Only admins listed in the config can use this feature.")
	}

	return true
}

// handleDMCommand processes commands sent in DMs
func handleDMCommand(s *discordgo.Session, m *discordgo.MessageCreate, input string, isAdmin bool) {
	inputvars := strings.SplitN(input+" ", " ", 2)
	commandName := strings.ToLower(inputvars[0])
	args := strings.TrimSpace(inputvars[1])

	// DM-exclusive commands
	switch commandName {
	case "verify":
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use the verify command.")
			return
		}
		handleVerifyCommand(s, m, args)
		return

	case "confirm":
		handleConfirmCommand(s, m, args)
		return

	case "unlink":
		handleUnlinkCommand(s, m)
		return

	case "status":
		handleStatusCommand(s, m)
		return

	case "c":
		// Open command (visible in chat and logs)
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use console commands.")
			return
		}
		handleOpenCommand(s, m, args)
		return

	case "sc":
		// Silent/hidden command (NOT visible in chat or logs)
		if !isAdmin {
			sendDMToChannel(s, m.ChannelID, "❌ Only admins can use silent commands.")
			return
		}
		handleSilentCommand(s, m, args)
		return

	case "help":
		sendDMHelp(s, m.ChannelID)
		return

	default:
		// Try to run as a regular command if admin
		if isAdmin {
			// Create a fake message for the command handler
			runDMRegularCommand(s, m, commandName, args)
		} else {
			sendDMToChannel(s, m.ChannelID, "❌ Unknown command. Use `"+support.Config.Prefix+"help` for available commands.")
		}
	}
}

// handleVerifyCommand initiates verification with a Factorio player
func handleVerifyCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	factorioUsername := strings.TrimSpace(args)
	if factorioUsername == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"verify <factorio_username>`\n\nThis will send a 6-digit verification code to the player in Factorio via a private whisper.")
		return
	}

	// Check if server is running
	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Generate verification code
	code, err := CreateVerificationRequest(m.Author.ID, factorioUsername)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ Failed to create verification request: "+err.Error())
		return
	}

	// Send the verification code to the Factorio player via silent whisper
	// Using RCON-style command that doesn't appear in logs
	// The /silent-command runs Lua code without logging to the console
	luaCommand := fmt.Sprintf(`/silent-command local p = game.get_player("%s"); if p and p.connected then p.print("[FactoCord Verification] Your code: %s - Enter this in Discord DM with: %sverify %s", {color={r=0.2,g=0.9,b=0.2}}) end`,
		escapeFactorioString(factorioUsername), code, support.Config.Prefix, code)

	support.Factorio.Send(luaCommand)

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Verification code sent to **%s** in Factorio!\n\nThe player should see a private message with a 6-digit code.\nOnce they tell you the code, use: `%sconfirm <code>`\n\n⏰ The code expires in 5 minutes.", factorioUsername, support.Config.Prefix))
}

// handleConfirmCommand confirms a verification code
func handleConfirmCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	code := strings.TrimSpace(args)
	if code == "" || len(code) != 6 {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"confirm <6-digit-code>`")
		return
	}

	success, factorioUsername, err := VerifyCode(m.Author.ID, code)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ Verification failed: "+err.Error())
		return
	}

	if !success {
		sendDMToChannel(s, m.ChannelID, "❌ Invalid verification code. Please try again.")
		return
	}

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Successfully verified! Your Discord account is now linked to Factorio player **%s**.\n\nYou can now use:\n• `%sc <command>` - Run commands (visible in game)\n• `%ssc <command>` - Run silent commands (hidden from logs)", factorioUsername, support.Config.Prefix, support.Config.Prefix))

	log.Printf("[DM] User %s (%s) verified as Factorio player %s", m.Author.Username, m.Author.ID, factorioUsername)
}

// handleUnlinkCommand removes the verification link
func handleUnlinkCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := UnlinkUser(m.Author.ID)
	if err != nil {
		sendDMToChannel(s, m.ChannelID, "❌ "+err.Error())
		return
	}

	sendDMToChannel(s, m.ChannelID, "✅ Your Discord account has been unlinked from your Factorio account.")
}

// handleStatusCommand shows verification status
func handleStatusCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)

	isAdmin := commands.CheckAdmin(m.Author.ID)
	adminStatus := "❌ Not an admin"
	if isAdmin {
		adminStatus = "✅ Admin"
	}

	verifyStatus := "❌ Not verified"
	if isVerified {
		verifyStatus = fmt.Sprintf("✅ Verified as **%s**", factorioUsername)
	}

	serverStatus := "❌ Server offline"
	if support.Factorio.IsRunning() {
		serverStatus = "✅ Server online"
	}

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("**Your Status:**\n• Admin: %s\n• Verification: %s\n• Server: %s", adminStatus, verifyStatus, serverStatus))
}

// handleOpenCommand runs a Factorio command (visible in chat and logs)
func handleOpenCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	if args == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"c <factorio_command>`\n\nThis command is visible in the Factorio chat and logs.")
		return
	}

	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Check if user is verified (required for running as a specific player)
	factorioUsername, isVerified := GetFactorioUsername(m.Author.ID)

	// Ensure command starts with /
	command := args
	if !strings.HasPrefix(command, "/") {
		command = "/" + command
	}

	// Run the command
	support.Factorio.Send(command)

	if isVerified {
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Command executed as **%s**: `%s`\n\n⚠️ This command was visible in the game chat and logs.", factorioUsername, command))
	} else {
		sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Command executed: `%s`\n\n⚠️ This command was visible in the game chat and logs.", command))
	}

	log.Printf("[DM] Open command from %s: %s", m.Author.Username, command)
}

// handleSilentCommand runs a Factorio command silently (NOT visible in chat or logs)
func handleSilentCommand(s *discordgo.Session, m *discordgo.MessageCreate, args string) {
	if args == "" {
		sendDMToChannel(s, m.ChannelID, "❌ Usage: `"+support.Config.Prefix+"sc <factorio_command_or_lua_code>`\n\nThis command is **hidden** from Factorio chat and logs.\n\nExamples:\n• `"+support.Config.Prefix+"sc game.print(\"Hello\")` - Run Lua code\n• `"+support.Config.Prefix+"sc game.speed = 2` - Change game speed")
		return
	}

	if !support.Factorio.IsRunning() {
		sendDMToChannel(s, m.ChannelID, "❌ The Factorio server is not running.")
		return
	}

	// Use /silent-command to run Lua code without logging
	// This ensures the command never appears in Factorio logs or chat
	command := "/silent-command " + args

	support.Factorio.Send(command)

	sendDMToChannel(s, m.ChannelID, fmt.Sprintf("✅ Silent command executed: `%s`\n\n🔒 This command was **hidden** from the game chat and logs.", args))

	// Note: We intentionally don't log the silent command content to maintain privacy
	log.Printf("[DM] Silent command from %s (content hidden for privacy)", m.Author.Username)
}

// runDMRegularCommand runs a regular bot command from DM context
func runDMRegularCommand(s *discordgo.Session, m *discordgo.MessageCreate, commandName, args string) {
	// Check if it's a help command
	if commandName == strings.ToLower("help") {
		sendDMHelp(s, m.ChannelID)
		return
	}

	// Find and execute the command
	for _, command := range commands.Commands {
		if strings.ToLower(command.Name) == commandName {
			// Check admin permission
			if command.Admin != nil && command.Admin(args) {
				if !commands.CheckAdmin(m.Author.ID) {
					sendDMToChannel(s, m.ChannelID, "❌ You don't have permission for this command.")
					return
				}
			}

			// Create a custom session handler that redirects output to DM
			// For now, just notify that the command was executed
			sendDMToChannel(s, m.ChannelID, fmt.Sprintf("⚙️ Executing command: `%s %s`\n\n_Note: Command output will appear in the main Factorio channel._", commandName, args))
			command.Command(s, args)
			return
		}
	}

	sendDMToChannel(s, m.ChannelID, "❌ Unknown command: `"+commandName+"`. Use `"+support.Config.Prefix+"help` for available commands.")
}

// sendDMHelp sends the DM help message
func sendDMHelp(s *discordgo.Session, channelID string) {
	embed := &discordgo.MessageEmbed{
		Type:        "rich",
		Color:       0x7289DA,
		Title:       "🤖 FactoCord DM Commands",
		Description: "You can control the Factorio server via Direct Messages.\n\n**Prefix:** `" + support.Config.Prefix + "`",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "🔐 Verification Commands",
				Value: fmt.Sprintf("`%sverify <player>` - Start verification with a Factorio player\n`%sconfirm <code>` - Confirm verification code\n`%sunlink` - Remove verification link\n`%sstatus` - Show your current status",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name: "🎮 Console Commands",
				Value: fmt.Sprintf("`%sc <command>` - Run a command (**visible** in game chat & logs)\n`%ssc <lua_code>` - Run silent Lua command (**hidden** from logs)",
					support.Config.Prefix, support.Config.Prefix),
			},
			{
				Name:  "📋 Regular Commands",
				Value: "All regular bot commands work here too:\n`server`, `save`, `kick`, `ban`, `unban`, `config`, `mod`, `mods`, `version`, `info`, `online`",
			},
			{
				Name: "ℹ️ Examples",
				Value: fmt.Sprintf("• `%sc /players` - List players (visible)\n• `%ssc game.print(\"Hello\")` - Silent Lua (hidden)\n• `%ssc game.speed = 2` - Change game speed (hidden)\n• `%sverify PlayerName` - Verify with player",
					support.Config.Prefix, support.Config.Prefix, support.Config.Prefix, support.Config.Prefix),
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "💡 Silent commands (sc) never appear in Factorio logs or chat",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Failed to send DM help: %v", err)
	}
}

// sendDMToChannel sends a message to a specific channel (usually DM)
func sendDMToChannel(s *discordgo.Session, channelID, message string) {
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("Failed to send DM: %v", err)
	}
}

// escapeFactorioString escapes a string for use in Factorio Lua commands
func escapeFactorioString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// NotifyDiscordUserOfVerification notifies a Discord user when someone verifies
func NotifyDiscordUserOfVerification(s *discordgo.Session, discordUserID, factorioUsername string) {
	// This can be called from the Factorio side when verification is complete
	channel, err := s.UserChannelCreate(discordUserID)
	if err != nil {
		log.Printf("Failed to create DM channel for notification: %v", err)
		return
	}

	sendDMToChannel(s, channel.ID, fmt.Sprintf("✅ Verification notification: Player **%s** has been linked to your account.", factorioUsername))
}
