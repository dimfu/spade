package tournament

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/spade/config"
	"github.com/dimfu/spade/database"
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/models"
	"github.com/google/uuid"
)

type TournamentRegisterHandler struct {
	Base             handler.BaseAdmin
	db               *sql.DB
	tournamentsModel *models.TournamentsModel
	playerModel      *models.PlayerModel
	attendeeModel    *models.AttendeeModel
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

func (h *TournamentRegisterHandler) players(inputs []string, s *discordgo.Session, tx *sql.Tx) []*models.Player {
	mode := config.GetEnv().ENV_MODE
	validPlayers := make([]*models.Player, 0, len(inputs))
	for _, player := range inputs {
		var validPl *models.Player
		discordId := player
		if strings.HasPrefix(player, "<@") {
			discordId = player[2 : len(player)-1]
		}

		p, err := h.playerModel.FindByDiscordId(discordId)
		if err != nil {
			if err == sql.ErrNoRows {
				name := p.Name
				user, err := s.User(discordId)
				if err != nil {
					log.Printf("error getting guild member: %v\n", err)
					if mode == "production" {
						continue
					}
				}

				if user != nil {
					name = user.Username
				}

				validPl = &models.Player{
					ID:        []uint8(uuid.New().String()),
					DiscordID: discordId,
					Name:      name,
				}
				err = h.playerModel.Insert(tx, validPl)
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

func (h *TournamentRegisterHandler) register(t *models.Tournament, p []*models.Player, sr bool, tx *sql.Tx) (int64, error) {
	var count int64
	if sr && len(p) == 1 {
		self, _ := h.attendeeModel.FindById(string(t.ID), string(p[0].ID))
		// ignoring the error cause that's what sigma does
		if self != nil {
			return 0, errors.New("You are already registered to this tournament.")
		}
	}

	for _, player := range p {
		q := `INSERT INTO attendees (tournament_id, player_id, current_seat) SELECT ?, ?, NULL
			  	WHERE NOT EXISTS (SELECT 1 FROM attendees WHERE tournament_id = ? AND player_id = ?)`

		result, err := tx.Exec(q, t.ID, player.ID, t.ID, player.ID)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Error inserting player %s as attendee: %v", player.ID, err)
		}
		rows, _ := result.RowsAffected()
		count += rows
	}

	return count, nil
}

func (h *TournamentRegisterHandler) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h.db = database.GetDB()
	h.tournamentsModel = models.NewTournamentsModel(h.db)
	h.playerModel = models.NewPlayerModel(h.db)
	h.attendeeModel = models.NewAttendeeModel(h.db)

	var inputs []string
	selfRegister := false

	data := i.ApplicationCommandData()

	if len(data.Options) > 0 {
		inputs = strings.Fields(data.Options[0].StringValue())
	}

	if len(inputs) > 0 {
		err := h.Base.HasPermit(s, i)
		if err != nil {
			handler.Respond("Insufficent permission to add other players to this tournament.", s, i, true)
			return
		}
	} else {
		inputs = append(inputs, i.Member.User.ID)
		selfRegister = true
	}

	tournamentId, err := h.tournamentsModel.GetTournamentIDInThread(i.ChannelID)
	if err != nil {
		handler.Respond(handler.ERR_GET_TOURNAMENT_IN_CHANNEL, s, i, true)
		return
	}

	t, err := h.tournamentsModel.GetById(string(tournamentId))
	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("error starting transaction %v", err)
		return
	}

	players := h.players(inputs, s, tx)
	regCount, err := h.register(t, players, selfRegister, tx)

	if err != nil {
		handler.Respond(err.Error(), s, i, true)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("error committing transaction: %v", err)
	}

	handler.Respond(fmt.Sprintf("Added %d players to the tournament.", regCount), s, i, true)
}
