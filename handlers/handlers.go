package handlers

import (
	"github.com/dimfu/spade/handlers/handler"
	"github.com/dimfu/spade/handlers/tournament"
)

var CommandHandlers = []handler.Command{
	&PingHandler{},
	&AdminHandler{},
	&tournament.TournamentCreateHandler{Base: handler.BaseAdmin{}},
	&tournament.TournamentDeleteHandler{Base: handler.BaseAdmin{}},
	&tournament.TournamentRegisterHandler{Base: handler.BaseAdmin{}},
	&ExportListHandler{Base: handler.BaseAdmin{}},
}

var ComponentHandlers = []handler.Component{
	&tournament.TournamentComponentHandler{Base: handler.BaseAdmin{}},
}

var ModalSubmitHandlers = []handler.Modal{
	&tournament.TournamentModalHandler{},
}
