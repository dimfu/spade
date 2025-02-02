package components

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type MatchupPayload struct {
	P1     string
	P2     string
	Winner *string
}

func MatchupEmbed(p MatchupPayload) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{}
	fields = append(fields, &discordgo.MessageEmbedField{Name: "Match Up", Value: fmt.Sprintf("%v vs %v", p.P1, p.P2)})

	if p.Winner != nil {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Result",
			Value:  fmt.Sprintf("%v took the win!", &p.Winner),
			Inline: true,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  "Round <count>",
		Fields: fields,
	}
}
