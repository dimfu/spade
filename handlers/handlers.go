package handlers

import (
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/handlers/tournament"
)

var CommandHandlers = []handler.Command{
	&PingHandler{},
	&AdminHandler{},
	&tournament.TournamentCreateHandler{Base: handler.GetBaseAdmin()},
	&tournament.TournamentDeleteHandler{Base: handler.GetBaseAdmin()},
	&tournament.TournamentRegisterHandler{Base: handler.GetBaseAdmin()},
	&tournament.ExportListHandler{Base: handler.GetBaseAdmin()},
	&tournament.SeedHandler{Base: handler.GetBaseAdmin()},
}

var ComponentHandlers = []handler.Component{
	&tournament.TournamentComponentHandler{Base: handler.GetBaseAdmin()},
}

var ModalSubmitHandlers = []handler.Modal{
	&tournament.TournamentModalHandler{},
}
