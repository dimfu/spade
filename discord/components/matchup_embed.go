package components

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/models"
)

type MatchupPayload struct {
	P1     models.AttendeeWithResult
	P2     models.AttendeeWithResult
	Winner *models.AttendeeWithResult
	Match  int
}

func MatchupEmbed(p MatchupPayload) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{}

	formatOpponent := func(player models.Player, isWinner *models.AttendeeWithResult) string {
		name := player.Name

		if player.DiscordID != "" {
			name = fmt.Sprintf("%s (<@%s>)", name, player.DiscordID)
		}

		if isWinner != nil {
			if isWinner.PlayerID == string(player.ID) {
				name = fmt.Sprintf("🏆 %s", name)
			} else {
				name = fmt.Sprintf("❌ ~~%s~~", name)
			}
		}

		return name
	}

	p1Name := formatOpponent(p.P1.Player, p.Winner)
	p2Name := formatOpponent(p.P2.Player, p.Winner)

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Opponent 1",
		Value: p1Name, Inline: true},
	)
	fields = append(fields, &discordgo.MessageEmbedField{
		Name:  "Opponent 2",
		Value: p2Name, Inline: true},
	)

	matchCount := p.Match
	if p.Winner != nil {
		matchCount--
	}

	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "spade",
			URL:     "https://www.github.com/dimfu/spade",
			IconURL: "https://cdn3.evostore.io/productimages/vow_api/l/sby23247_01.jpg",
		},
		Title:  fmt.Sprintf("Tournament Match #%d", matchCount),
		Fields: fields,
	}
}
