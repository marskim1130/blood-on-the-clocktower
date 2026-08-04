package ws

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marskim1130/blood-on-the-clocktower/internal/game"
)

// ClientGamePhase accepts the numeric protocol enum and compact JSON names.
type ClientGamePhase game.GamePhase

func (p ClientGamePhase) GamePhase() game.GamePhase { return game.GamePhase(p) }

func (p *ClientGamePhase) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*p = ClientGamePhase(game.GamePhaseUnspecified)
		return nil
	}
	var numeric int
	if err := json.Unmarshal(raw, &numeric); err == nil {
		*p = ClientGamePhase(game.GamePhase(numeric))
		return nil
	}
	var named string
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}
	phase, ok := parseClientGamePhase(named)
	if !ok {
		return fmt.Errorf("unknown game phase %q", named)
	}
	*p = ClientGamePhase(phase)
	return nil
}

func parseClientGamePhase(value string) (game.GamePhase, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "unspecified", "0":
		return game.GamePhaseUnspecified, true
	case "setup", "1":
		return game.GamePhaseSetup, true
	case "day", "2":
		return game.GamePhaseDay, true
	case "night", "3":
		return game.GamePhaseNight, true
	case "voting", "4":
		return game.GamePhaseVoting, true
	case "finished", "5":
		return game.GamePhaseFinished, true
	default:
		return game.GamePhaseUnspecified, false
	}
}

// ClientTeam accepts the numeric protocol enum and compact JSON names.
type ClientTeam game.Team

func (t ClientTeam) Team() game.Team { return game.Team(t) }

func (t *ClientTeam) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*t = ClientTeam(game.TeamUnspecified)
		return nil
	}
	var numeric int
	if err := json.Unmarshal(raw, &numeric); err == nil {
		*t = ClientTeam(game.Team(numeric))
		return nil
	}
	var named string
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}
	team, ok := parseClientTeam(named)
	if !ok {
		return fmt.Errorf("unknown team %q", named)
	}
	*t = ClientTeam(team)
	return nil
}

func parseClientTeam(value string) (game.Team, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "good", "1":
		return game.TeamGood, true
	case "evil", "2":
		return game.TeamEvil, true
	case "unspecified", "0":
		return game.TeamUnspecified, true
	default:
		return game.TeamUnspecified, false
	}
}

// ClientDeathCause accepts the numeric protocol enum and compact JSON names.
type ClientDeathCause game.DeathCause

func (c ClientDeathCause) DeathCause() game.DeathCause { return game.DeathCause(c) }

func (c *ClientDeathCause) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*c = ClientDeathCause("")
		return nil
	}
	var numeric int
	if err := json.Unmarshal(raw, &numeric); err == nil {
		cause, ok := parseClientDeathCause(fmt.Sprintf("%d", numeric))
		if !ok {
			return fmt.Errorf("unknown death cause %d", numeric)
		}
		*c = ClientDeathCause(cause)
		return nil
	}
	var named string
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}
	cause, ok := parseClientDeathCause(named)
	if !ok {
		return fmt.Errorf("unknown death cause %q", named)
	}
	*c = ClientDeathCause(cause)
	return nil
}

func parseClientDeathCause(value string) (game.DeathCause, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "unspecified", "0":
		return "", true
	case "execution", "1":
		return game.DeathCauseExecution, true
	case "night_kill", "night-kill", "night kill", "2":
		return game.DeathCauseNightKill, true
	case "ability", "3":
		return game.DeathCauseAbility, true
	default:
		return "", false
	}
}

func (message ClientMessage) targetPlayerID() string { return message.TargetPlayerID }
