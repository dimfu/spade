package discord

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/config"
	"github.com/dimfu/spade/handlers"
)

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

	// register commands
	for _, handler := range handlers.CommandHandlers {
		cmd := handler.Command()
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("error creating command %s: %v", cmd.Name, err)
			continue
		}
	}

	// listens to which command is being used, and do the handler
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			for _, handler := range handlers.CommandHandlers {
				if handler.Command().Name == i.ApplicationCommandData().Name {
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
