package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type PingHandler struct{}

func (p *PingHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Ping to get ponged",
	}
}

func (p *PingHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! %vms", s.HeartbeatLatency().Milliseconds()),
		},
	})
}
