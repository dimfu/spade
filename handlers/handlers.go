package handlers

import (
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/handlers/tournament"
)

var CommandHandlers = []base.Command{
	&PingHandler{},
	&AdminHandler{},
	&tournament.TournamentCreateHandler{Base: base.GetBaseAdmin()},
	&tournament.TournamentDeleteHandler{Base: base.GetBaseAdmin()},
	&tournament.TournamentRegisterHandler{Base: base.GetBaseAdmin()},
	&tournament.ExportListHandler{Base: base.GetBaseAdmin()},
	&tournament.SeedHandler{Base: base.GetBaseAdmin()},
	&tournament.StartHandler{Base: *base.GetBaseAdmin()},
}

var ComponentHandlers = []base.Component{
	&tournament.TournamentComponentHandler{Base: base.GetBaseAdmin()},
}

var ModalSubmitHandlers = []base.Modal{
	&tournament.TournamentModalHandler{},
}
