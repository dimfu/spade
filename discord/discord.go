package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/config"
	"github.com/dimfu/spade/handlers"
	"github.com/dimfu/spade/handlers/base"
)

func ensureRole(dg *discordgo.Session, gid string) (*discordgo.Role, error) {
	st, err := dg.GuildRoles(gid)

	if err != nil {
		return nil, err
	}

	for _, role := range st {
		if role.Name == "Tournament Manager" {
			return role, nil
		}
	}

	r, err := dg.GuildRoleCreate(
		gid,
		&discordgo.RoleParams{
			Name: "Tournament Manager",
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return r, nil
}

func Init(ctx context.Context) {
	config := config.GetEnv()
	dg, err := discordgo.New("Bot " + config.DISCORD_BOT_TOKEN)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = dg.Open()
	if err != nil {
		log.Fatalf("error opening connection with discord: %v", err)
	}
	defer dg.Close()

	if err != nil {
		log.Fatalf("error while initializing role: %v", err)
	}

	// register commands
	for _, handler := range handlers.CommandHandlers {
		cmd := handler.Command()
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("error creating command %s: %v", cmd.Name, err)
			continue
		}
	}

	// create role when bot joins new guild
	dg.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		_, err := ensureRole(s, g.ID)
		if err != nil {
			log.Printf("Failed to create role in guild %s: %v", g.ID, err)
		}
	})

	// listens to which command is being used, and do the handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			for _, handler := range handlers.CommandHandlers {
				if handler.Command().Name == i.ApplicationCommandData().Name {
					if hWithCtx, ok := handler.(base.CommandWithCtx); ok {
						hWithCtx.WithCtx(ctx)
					}
					handler.Handler(dg, i)
					return
				}
			}
		case discordgo.InteractionMessageComponent:
			for _, handler := range handlers.ComponentHandlers {
				if strings.HasPrefix(i.MessageComponentData().CustomID, handler.Name()) {
					handler.Handler(dg, i)
					return
				}
			}
		case discordgo.InteractionModalSubmit:
			for _, handler := range handlers.ModalSubmitHandlers {
				if strings.HasPrefix(i.ModalSubmitData().CustomID, handler.Name()) {
					handler.Handler(dg, i)
					return
				}
			}
		}
	})

	if err != nil {
		log.Fatalf("error creating slash commands: %v", err)
	}

	log.Println("bot is now running")
	<-ctx.Done()
}
