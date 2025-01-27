package base

import (
	"errors"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	ERR_INTERNAL_ERROR            = "Something went wrong while executing this instruction"
	ERR_CREATING_TOURNAMENT       = "Something went wrong while creating tournament"
	ERR_GENERATE_BRACKET          = "Error occured when generating tournament bracket template"
	ERR_GET_TOURNAMENT            = "Cannot find this tournament record"
	ERR_GET_TOURNAMENT_IN_CHANNEL = "Cannot find tournament in this channel, make sure to run this command inside the tournament channel"
	ERR_GET_TOURNAMENT_TYPES      = "Cannot get the list of tournament types"
)

var (
	instance *BaseAdmin
	once     sync.Once
)

type Command interface {
	Command() *discordgo.ApplicationCommand
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Component interface {
	Name() string
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Modal interface {
	Name() string
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type BaseAdmin struct {
	Count int
}

func GetBaseAdmin() *BaseAdmin {
	once.Do(func() {
		instance = &BaseAdmin{}
	})
	return instance
}

func Respond(r string, s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral bool) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: r,
		},
	}

	if ephemeral {
		response.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	s.InteractionRespond(i.Interaction, response)
}

func (h *BaseAdmin) HasPermit(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	return errors.New("Insufficent permission to use this command")
}
