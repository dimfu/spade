package tournament

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
	"github.com/google/uuid"
)

type TournamentRegisterHandler struct {
	Base handler.BaseAdmin
	db   *sql.DB
}

func (h *TournamentRegisterHandler) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "register",
		Description: "Register a user to current tournament",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "players",
				Description: "Player(s) to be registered",
				Required:    false,
			},
		},
	}
}

func (h *TournamentRegisterHandler) players(inputs []string, s *discordgo.Session, i *discordgo.InteractionCreate, tx *sql.Tx) []*models.Player {
	pm := models.NewPlayerModel(h.db)
	validPlayers := make([]*models.Player, 0, len(inputs))
	for _, player := range inputs {
		discordId := player[2 : len(player)-1]
		dMember, err := s.GuildMember(i.GuildID, discordId)
		if err != nil {
			log.Printf("error getting guild member: %v\n", err)
			continue
		}

		var validPl *models.Player

		p, err := pm.FindByDiscordId(discordId)
		if err != nil {
			if err == sql.ErrNoRows {
				validPl = &models.Player{
					ID:        []uint8(uuid.New().String()),
					DiscordID: discordId,
					Name:      dMember.User.Username,
				}
				err := pm.Insert(tx, validPl)
				if err != nil {
					log.Println(err.Error())
					continue
				}
			} else {
				log.Printf("something went wrong while getting player %v\n", err)
				continue
			}
		} else {
			validPl = p
		}

		validPlayers = append(validPlayers, validPl)
	}

	return validPlayers
}

func (h *TournamentRegisterHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := h.Base.HasPermit(s, i)
	if err != nil {
		log.Println(err)
		return
	}

	h.db = database.GetDB()
	tm := models.NewTournamentsModel(h.db)

	data := i.ApplicationCommandData()
	inputs := strings.Fields(data.Options[0].StringValue())
	log.Println(inputs)

	tournamentId, err := tm.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	t, err := tm.GetById(string(tournamentId))
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("error starting transaction %v", err)
		return
	}

	var registeredPlayers int64
	players := h.players(inputs, s, i, tx)
	for _, player := range players {
		q := `INSERT INTO attendees (tournament_id, player_id, current_seat) SELECT ?, ?, NULL
			  	WHERE NOT EXISTS (SELECT 1 FROM attendees WHERE tournament_id = ? AND player_id = ?)`

		result, err := tx.Exec(q, t.ID, player.ID, t.ID, player.ID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("error inserting player %s as attendee: %v", player.ID, err)
			return
		}
		count, _ := result.RowsAffected()
		registeredPlayers += count
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("error committing transaction: %v", err)
	}

	handler.Respond(fmt.Sprintf("Added %d players to the tournament.", registeredPlayers), s, i, true)
}
