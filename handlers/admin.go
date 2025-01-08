package handlers

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type AdminHandler struct{}

var defaultPermission = int64(discordgo.PermissionAdministrator)

func (p *AdminHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     "admin",
		Description:              "Manage permission to manage spade tournaments",
		DefaultMemberPermissions: &defaultPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add tournament admin role to a user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "target",
						Description: "Target User",
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove tournament admin role to a user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "target",
						Description: "Target User",
						Required:    true,
					},
				},
			},
		},
	}
}

func (p *AdminHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if len(data.Options) <= 0 {
		log.Println("empty options")
		return
	}

	cmd := i.ApplicationCommandData().Options
	subcmd := data.Options[0]
	var st *discordgo.User

	if len(subcmd.Options) > 0 {
		payload := subcmd.Options[0].StringValue()
		usrid := payload[2 : len(payload)-1]
		u, err := s.User(usrid)
		if err != nil {
			respond("Cannot add invalid user, use @user to properly target user", s, i, true)
		}
		st = u
	}

	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Println(err.Error())
		return
	}

	var tm *discordgo.Role
	for _, role := range roles {
		if role.Name == "Tournament Manager" {
			tm = role
		}
	}

	if tm == nil {
		respond("Can't find tournament role", s, i, false)
		return
	}

	ret := ""

	switch cmd[0].Name {
	case "add":
		if err := s.GuildMemberRoleAdd(i.GuildID, st.ID, tm.ID); err != nil {
			ret = err.Error()
			break
		}
		ret = fmt.Sprintf("<@%s> is now tournament manager", st.ID)
	case "remove":
		if err := s.GuildMemberRoleRemove(i.GuildID, st.ID, tm.ID); err != nil {
			ret = err.Error()
			break
		}
		ret = fmt.Sprintf("<@%s> is no longer tournament manager", st.ID)
	default:
	}

	respond(ret, s, i, false)
}
