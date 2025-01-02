package handlers

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler interface {
	Command() *discordgo.ApplicationCommand
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type BaseAdminHandler struct{}

func (bah *BaseAdminHandler) HasPermit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := i.Member

	// skips check if user has asdmin access
	if user.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		return nil
	}

	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return err
	}

	var tm *discordgo.Role
	for _, role := range roles {
		if role.Name == "Tournament Manager" {
			tm = role
		}
	}

	if tm == nil {
		return errors.New("Can't find tournament role")
	}

	for _, ur := range user.Roles {
		if ur == tm.ID {
			return nil
		}
	}

	return errors.New("Insufficent permission to create tournament")
}

var CommandHandlers = []CommandHandler{
	&PingHandler{},
	&AdminHandler{},
	&TournamentCreateHandler{base: BaseAdminHandler{}},
}
