package utils

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/models"
	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// HandleResearchCommand zeigt Forschungsfortschritt und Tech-Tree Informationen
// Nutzung:
//
//	$research - Kurzübersicht
//	$research status - Kurzübersicht mit Statistiken
//	$research tree - Kompletter Tech-Tree
//	$research queue - Forschungs-Queue
//	$research <name> - Details zu spezifischer Technologie
func HandleResearchCommand(s *discordgo.Session, args string) {
	if !models.HasTechTreeData() {
		support.Send(s, "❌ Tech-Tree Daten noch nicht geladen. Bitte warten Sie...")
		return
	}

	tree := models.GetTechTree()
	if tree == nil {
		support.Send(s, "❌ Tech-Tree Daten sind nicht verfügbar")
		return
	}

	args = strings.ToLower(strings.TrimSpace(args))

	// Wenn leer oder "status": Zeige Kurzübersicht
	if args == "" || args == "status" {
		displayResearchStatus(s, tree)
		return
	}

	// Parse Subcommands
	parts := strings.Fields(args)
	subcommand := parts[0]

	switch subcommand {
	case "tree":
		displayResearchTree(s, tree)
	case "queue":
		displayResearchQueue(s, tree)
	case "available":
		displayAvailableTechs(s, tree)
	case "all":
		displayAllTechs(s, tree)
	case "current":
		displayCurrentResearch(s, tree)
	default:
		// Versuche als Tech-Name zu interpretieren
		displayResearchDetails(s, tree, args)
	}
}

// displayResearchStatus zeigt eine Kurzübersicht des Forschungsfortschritts
func displayResearchStatus(s *discordgo.Session, tree *models.TechTree) {
	if tree == nil || tree.Stats == nil {
		support.Send(s, "❌ Keine Daten verfügbar")
		return
	}

	stats := tree.Stats
	currentName := "Keine"
	currentProgress := "0%"

	if tree.Current != nil {
		currentName = tree.Current.Name
		// Berechne Fortschritt basierend auf Cost
		if tree.Current.Cost > 0 {
			// Diese Berechnung müsste von Factorio-Seite kommen
			currentProgress = "Lädt..."
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🔬 Forschungs-Übersicht",
		Color:       0x5865F2,
		Description: "Aktuelle Forschungs-Statistik des Servers",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "✅ Erforschte Technologien",
				Value:  fmt.Sprintf("**%d**/%d", stats.ResearchedCount, stats.TotalTechs),
				Inline: true,
			},
			{
				Name:   "⏳ Aktuelle Forschung",
				Value:  fmt.Sprintf("`%s`\n%s", currentName, currentProgress),
				Inline: true,
			},
			{
				Name:   "🟨 Direkt verfügbar",
				Value:  fmt.Sprintf("**%d**", stats.AvailableDirectCount),
				Inline: true,
			},
			{
				Name:   "📋 Nach Forschung verfügbar",
				Value:  fmt.Sprintf("**%d**", stats.AvailableAfterCount),
				Inline: true,
			},
			{
				Name:   "❓ Nicht verfügbar",
				Value:  fmt.Sprintf("**%d**", stats.UnavailableCount),
				Inline: true,
			},
			{
				Name:   "⏲️ In Schlange",
				Value:  fmt.Sprintf("**%d**", stats.QueueLength),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Aktualisiert: %s • Nutze %sresearch tree für Details",
				tree.LastUpdate.Format("15:04:05"), support.Config.Prefix),
		},
	}

	support.SendEmbed(s, embed)
}

// displayResearchTree zeigt den kompletten Tech-Tree mit Struktur
func displayResearchTree(s *discordgo.Session, tree *models.TechTree) {
	if tree == nil || tree.Stats == nil {
		support.Send(s, "❌ Keine Daten verfügbar")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🌳 Technologie-Baum",
		Color:       0x5865F2,
		Description: "Übersicht aller verfügbaren und erforschten Technologien",
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	// ✅ Erforschte Technologien
	researched := models.GetResearchesByState(models.ResearchedState)
	researchedText := formatTechList(researched, 20)
	if researchedText == "" {
		researchedText = "Keine"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("✅ Erforschte Technologien (%d)", len(researched)),
		Value:  researchedText,
		Inline: false,
	})

	// ⏳ Aktuelle Forschung
	if tree.Current != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⏳ Aktuell in Forschung",
			Value:  fmt.Sprintf("`%s` • Kosten: **%d**", tree.Current.Name, tree.Current.Cost),
			Inline: false,
		})
	}

	// 🟨 Direkt verfügbar
	available := models.GetResearchesByState(models.AvailableDirectState)
	availableText := formatTechList(available, 15)
	if availableText == "" {
		availableText = "Keine"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("🟨 Direkt verfügbar (%d)", len(available)),
		Value:  availableText,
		Inline: false,
	})

	// 📋 Nach Forschung verfügbar
	afterResearch := models.GetResearchesByState(models.AvailableAfterState)
	afterText := formatTechList(afterResearch, 10)
	if afterText == "" {
		afterText = "Keine"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("📋 Nach Forschung verfügbar (%d)", len(afterResearch)),
		Value:  afterText,
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Zuletzt aktualisiert: %s", tree.LastUpdate.Format("15:04:05")),
	}

	support.SendEmbed(s, embed)
}

// displayResearchQueue zeigt die Forschungs-Queue
func displayResearchQueue(s *discordgo.Session, tree *models.TechTree) {
	if tree == nil {
		support.Send(s, "❌ Keine Daten verfügbar")
		return
	}

	queue := models.GetResearchQueue()

	if len(queue) == 0 {
		support.Send(s, "📭 Keine Forschungs-Queue aktiv\n\n"+
			"Die Forschung läuft entweder aktuell oder die Queue ist leer.")
		return
	}

	var queueText strings.Builder
	queueText.WriteString("```\n")
	for i, item := range queue {
		if i >= 10 { // Limit auf 10 für Lesbarkeit
			queueText.WriteString(fmt.Sprintf("\n... und %d weitere in der Queue\n", len(queue)-i))
			break
		}
		emoji := "⏳"
		if i == 0 {
			emoji = "🔜"
		}
		queueText.WriteString(fmt.Sprintf("%s %d. %s\n", emoji, i+1, item.Name))
	}
	queueText.WriteString("```")

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Forschungs-Queue",
		Description: queueText.String(),
		Color:       0xFFA500,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Insgesamt %d Technologien in der Queue", len(queue)),
		},
	}

	support.SendEmbed(s, embed)
}

// displayAvailableTechs zeigt alle direkt verfügbaren Technologien
func displayAvailableTechs(s *discordgo.Session, tree *models.TechTree) {
	available := models.GetResearchesByState(models.AvailableDirectState)

	if len(available) == 0 {
		support.Send(s, "❌ Keine direkt verfügbaren Technologien")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🟨 Verfügbare Technologien (%d)", len(available)),
		Color:       0xFFA500,
		Description: "Diese Technologien können sofort erforscht werden",
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	for i, tech := range available {
		if i >= 20 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("... und %d weitere", len(available)-i),
				Value:  "Nutze `" + support.Config.Prefix + "research tree` für alle",
				Inline: false,
			})
			break
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   tech.Name,
			Value:  fmt.Sprintf("Kosten: **%d**", tech.Cost),
			Inline: true,
		})
	}

	support.SendEmbed(s, embed)
}

// displayAllTechs zeigt alle Technologien mit ihrem Status
func displayAllTechs(s *discordgo.Session, tree *models.TechTree) {
	if tree == nil || tree.AllTechs == nil {
		support.Send(s, "❌ Keine Daten verfügbar")
		return
	}

	// Gruppiere nach State
	var researched, current, direct, after, unavailable strings.Builder

	researched.WriteString("```\n")
	current.WriteString("```\n")
	direct.WriteString("```\n")
	after.WriteString("```\n")
	unavailable.WriteString("```\n")

	rCount, cCount, dCount, aCount, uCount := 0, 0, 0, 0, 0

	for _, tech := range tree.AllTechs {
		switch tech.State {
		case models.ResearchedState:
			if rCount < 5 {
				researched.WriteString(fmt.Sprintf("✅ %s\n", tech.Name))
				rCount++
			}
		case models.CurrentState:
			if cCount < 5 {
				current.WriteString(fmt.Sprintf("⏳ %s\n", tech.Name))
				cCount++
			}
		case models.AvailableDirectState:
			if dCount < 5 {
				direct.WriteString(fmt.Sprintf("🟨 %s\n", tech.Name))
				dCount++
			}
		case models.AvailableAfterState:
			if aCount < 5 {
				after.WriteString(fmt.Sprintf("📋 %s\n", tech.Name))
				aCount++
			}
		case models.UnavailableState:
			if uCount < 5 {
				unavailable.WriteString(fmt.Sprintf("❌ %s\n", tech.Name))
				uCount++
			}
		}
	}

	researched.WriteString("```")
	current.WriteString("```")
	direct.WriteString("```")
	after.WriteString("```")
	unavailable.WriteString("```")

	embed := &discordgo.MessageEmbed{
		Title:       "📊 Alle Technologien",
		Color:       0x5865F2,
		Description: "Übersicht aller Technologien (Beispiele)",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   fmt.Sprintf("✅ Erforschte (%d)", models.GetResearchStats().ResearchedCount),
				Value:  researched.String(),
				Inline: true,
			},
			{
				Name:   fmt.Sprintf("🟨 Verfügbar (%d)", models.GetResearchStats().AvailableDirectCount),
				Value:  direct.String(),
				Inline: true,
			},
			{
				Name:   fmt.Sprintf("📋 Nach Forschung (%d)", models.GetResearchStats().AvailableAfterCount),
				Value:  after.String(),
				Inline: true,
			},
		},
	}

	support.SendEmbed(s, embed)
}

// displayCurrentResearch zeigt die aktuelle Forschung
func displayCurrentResearch(s *discordgo.Session, tree *models.TechTree) {
	if tree == nil || tree.Current == nil {
		support.Send(s, "❌ Keine aktuelle Forschung läuft")
		return
	}

	current := tree.Current
	prereqText := "Keine"
	if len(current.Prerequisites) > 0 {
		prereqText = strings.Join(current.Prerequisites, ", ")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⏳ %s", current.Name),
		Color:       0xFFAA00,
		Description: "Aktuelle Forschung",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Kosten",
				Value:  fmt.Sprintf("**%d** Forschungspunkte", current.Cost),
				Inline: true,
			},
			{
				Name:   "Level",
				Value:  fmt.Sprintf("**%d**", current.Level),
				Inline: true,
			},
			{
				Name:   "Voraussetzungen",
				Value:  prereqText,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Aktualisiert: %s", tree.LastUpdate.Format("15:04:05")),
		},
	}

	support.SendEmbed(s, embed)
}

// displayResearchDetails zeigt Details zu einer spezifischen Technologie
func displayResearchDetails(s *discordgo.Session, tree *models.TechTree, searchName string) {
	// Suche nach exaktem Match oder Substring
	var found *models.Research
	searchLower := strings.ToLower(searchName)

	for _, tech := range tree.AllTechs {
		if strings.ToLower(tech.Name) == searchLower {
			found = tech
			break
		}
	}

	// Wenn kein exakter Match, suche nach Substring
	if found == nil {
		for _, tech := range tree.AllTechs {
			if strings.Contains(strings.ToLower(tech.Name), searchLower) {
				found = tech
				break
			}
		}
	}

	if found == nil {
		support.Send(s, fmt.Sprintf("❌ Technologie '%s' nicht gefunden", searchName))
		return
	}

	// Bestimme Status-Emoji
	statusEmoji := "❓"
	statusColor := 0x808080
	switch found.State {
	case models.ResearchedState:
		statusEmoji = "✅"
		statusColor = 0x57F287
	case models.CurrentState:
		statusEmoji = "⏳"
		statusColor = 0xFFAA00
	case models.AvailableDirectState:
		statusEmoji = "🟨"
		statusColor = 0xFFA500
	case models.AvailableAfterState:
		statusEmoji = "📋"
		statusColor = 0xFFC700
	case models.UnavailableState:
		statusEmoji = "❌"
		statusColor = 0xFF0000
	}

	prereqText := "Keine"
	if len(found.Prerequisites) > 0 {
		prereqText = "`" + strings.Join(found.Prerequisites, "`, `") + "`"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", statusEmoji, found.Name),
		Color:       statusColor,
		Description: fmt.Sprintf("Level: **%d**", found.Level),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  fmt.Sprintf("`%s`", found.State),
				Inline: true,
			},
			{
				Name:   "Kosten",
				Value:  fmt.Sprintf("**%d**", found.Cost),
				Inline: true,
			},
			{
				Name:   "Voraussetzungen",
				Value:  prereqText,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Aktualisiert: %s", tree.LastUpdate.Format("15:04:05")),
		},
	}

	support.SendEmbed(s, embed)
}

// formatTechList formatiert eine Liste von Technologien für die Anzeige
func formatTechList(techs []*models.Research, maxItems int) string {
	if len(techs) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("```\n")

	for i, tech := range techs {
		if i >= maxItems {
			result.WriteString(fmt.Sprintf("... und %d weitere\n", len(techs)-i))
			break
		}
		result.WriteString(fmt.Sprintf("• %s\n", tech.Name))
	}

	result.WriteString("```")
	return result.String()
}

// ResearchDoc ist die Dokumentation für den Research-Command
var ResearchDoc = support.CommandDoc{
	Name: "research",
	Usage: "$research\n" +
		"$research status\n" +
		"$research tree\n" +
		"$research queue\n" +
		"$research available\n" +
		"$research current\n" +
		"$research <name>",
	Doc: "Zeige Forschungsfortschritt und Tech-Tree Informationen. " +
		"Nutze Subcommands für verschiedene Ansichten oder den Namen einer Technologie für Details.",
	Subcommands: []support.CommandDoc{
		{
			Name: "status",
			Doc:  "Kurzübersicht mit Statistiken",
		},
		{
			Name: "tree",
			Doc:  "Kompletter Tech-Tree mit allen Technologien",
		},
		{
			Name: "queue",
			Doc:  "Forschungs-Queue anzeigen",
		},
		{
			Name: "available",
			Doc:  "Alle direkt verfügbaren Technologien",
		},
		{
			Name: "current",
			Doc:  "Details zur aktuellen Forschung",
		},
		{
			Name: "all",
			Doc:  "Übersicht aller Technologien",
		},
	},
}
