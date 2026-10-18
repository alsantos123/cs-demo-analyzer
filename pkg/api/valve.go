package api

import (
	"github.com/akiver/cs-demo-analyzer/pkg/api/constants"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msgs2"
)

func createValveAnalyzer(analyzer *Analyzer) {
	parser := analyzer.parser
	match := analyzer.match
	analyzer.matchStarted = parser.GameState().IsMatchStarted

	parser.RegisterNetMessageHandler(func(srvInfo *msgs2.CSVCMsg_ServerInfo) {
		match.gameModeStr = constants.GameModeStr(srvInfo.GameSessionConfig.GetGamemode())
	})

	parser.RegisterEventHandler(func(event events.MatchStart) {
		currentRound := analyzer.currentRound
		currentRound.StartFrame = parser.CurrentFrame()
		currentRound.StartTick = analyzer.currentTick()
		analyzer.updateTeamNames()
	})

	parser.RegisterEventHandler(func(event events.MatchStartedChanged) {
		isMatchStarted := !event.OldIsStarted && event.NewIsStarted
		if isMatchStarted {
			currentRound := analyzer.currentRound
			currentRound.TeamASide = *match.TeamA.CurrentSide
			currentRound.TeamBSide = *match.TeamB.CurrentSide
		}
	})

	parser.RegisterEventHandler(func(event events.GameHalfEnded) {
		analyzer.isFirstRoundOfHalf = true
	})

	parser.RegisterEventHandler(analyzer.defaultRoundFreezetimeChangedHandler)

	// A player keeps the Team he was FIRST seen on (registerPlayer -> NewPlayer):
	// the side he held during warm-up, or during a knife round whose post-knife
	// swap the GamePhaseChanged handler ignores (match not started). Nothing in
	// this analyzer re-syncs him afterwards — processMatchStart (the only caller
	// of updatePlayersCurrentTeam) is never invoked for the Valve source — so he
	// stays on the wrong team for the whole match: one late/pre-swapped player
	// lands on the opponent's roster (a 6v4 scoreboard), or a whole roster
	// inverts against the team names and scores after a knife-round swap.
	// Re-sync everyone to the side he actually holds at every round start.
	// Source 2 only: there the side-switch entity update precedes round_start
	// (see the GamePhaseChanged comment), so TeamA/TeamB.CurrentSide is already
	// swapped when this runs; on CS:GO demos round_start fires first and this
	// would flip every player onto the wrong team at each half.
	parser.RegisterEventHandler(func(event events.RoundStart) {
		if analyzer.isSource2 {
			analyzer.updatePlayersCurrentTeam()
		}
	})

	parser.RegisterEventHandler(analyzer.defaultRoundStartHandler)

	parser.RegisterEventHandler(analyzer.defaultRoundEndOfficiallyHandler)

	parser.RegisterEventHandler(analyzer.defaultAnnouncementWinPanelMatchHandler)
}
