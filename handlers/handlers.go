package handlers

import (
	"github.com/dimfu/spade/handlers/base"
	"github.com/dimfu/spade/handlers/queue"
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
	&tournament.StartHandler{
		Base:       base.GetBaseAdmin(),
		MatchQueue: queue.GetMatchQueue(),
	},
	&tournament.RestartTournamentHandler{
		Base:       base.GetBaseAdmin(),
		MatchQueue: queue.GetMatchQueue(),
	},
}

var ComponentHandlers = []base.Component{
	&tournament.TournamentComponentHandler{Base: base.GetBaseAdmin(), MatchQueue: queue.GetMatchQueue()},
}

var ModalSubmitHandlers = []base.Modal{
	&tournament.TournamentModalHandler{},
}
