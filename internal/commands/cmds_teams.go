package commands

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// the default largest allowable skill gap between the strongest and weakest created teams.
const defaultTeamsMaxSkillGap = float64(20)
const teamGenTimeLimit = 100 * time.Millisecond

func cmdTeams(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	numTeams := int(options[0].IntValue())
	maxSkillGap := defaultTeamsMaxSkillGap
	if len(options) > 1 {
		maxSkillGap = float64(options[1].IntValue())
	}

	players, err := getPlaying(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	if len(players) < numTeams {
		interactionRespondf(session, interaction, "%d players not enough to make %d teams", len(players), numTeams)
		return
	}

	// ensure that all playing users have a skill rank set
	unsetPlayers := []string{}
	for _, player := range players {
		if player.Skill == -1 {
			unsetPlayers = append(unsetPlayers, player.Name)
		}
	}
	if len(unsetPlayers) > 0 {
		str := "Playing users with ranks undefined:"
		for _, name := range unsetPlayers {
			str = fmt.Sprintf("%s\n\t%s", str, name)
		}
		interactionRespond(session, interaction, str)
		return
	}

	teams := createTeams(players, numTeams, maxSkillGap, teamGenTimeLimit)

	if teams.skillGap <= maxSkillGap {
		str := fmt.Sprintf("Teams found:%s", teams.String())
		interactionRespond(session, interaction, str)
	} else {
		str := fmt.Sprintf("No valid team. Best option:%s", teams.String())
		interactionRespond(session, interaction, str)
	}
}

func createTeams(
	players []Player,
	numTeams int,
	maxSkillGap float64,
	timeLimit time.Duration,
) Teams {
	// put an upper time limit on shuffling for good teams
	deadline := time.Now().Add(timeLimit)

	// loop variables
	bestArrangementDiff := float64(100)
	bestArrangement := make([]int, len(players))
	bestAverages := make([]float64, numTeams)
	arrangement := make([]int, len(players))
	averages := make([]float64, numTeams)
	for i := range arrangement {
		arrangement[i] = i
	}
	teamSizes := make([]int, numTeams)
	for teamIdx := 0; teamIdx < numTeams; teamIdx++ {
		if len(players)%numTeams > teamIdx {
			teamSizes[teamIdx] = len(players)/numTeams + 1
		} else {
			teamSizes[teamIdx] = len(players) / numTeams
		}
	}

	// shuffle until time runs our or a good solution is found
	for {
		rand.Shuffle(len(arrangement), func(i, j int) {
			arrangement[i], arrangement[j] = arrangement[j], arrangement[i]
		})

		maxAverage := float64(0)
		minAverage := float64(100)
		playerIdx := 0
		for teamIdx := 0; teamIdx < numTeams; teamIdx++ {
			sum := 0
			for teammateIdx := 0; teammateIdx < teamSizes[teamIdx]; teammateIdx++ {
				sum += players[arrangement[playerIdx]].Skill
				playerIdx++
			}
			average := float64(sum) / float64(teamSizes[teamIdx])
			averages[teamIdx] = average
			if average > float64(maxAverage) {
				maxAverage = average
			}
			if average < minAverage {
				minAverage = average
			}
		}
		averageDiff := maxAverage - minAverage
		if averageDiff < bestArrangementDiff {
			bestArrangementDiff = averageDiff
			copy(bestArrangement, arrangement)
			copy(bestAverages, averages)
		}
		if averageDiff <= maxSkillGap {
			break
		}
		if time.Until(deadline) < 0 {
			break
		}
	}

	playerIdx := 0
	teams := Teams{
		skillGap: bestArrangementDiff,
		teams:    make([]*Team, numTeams),
	}
	for teamIdx := 0; teamIdx < numTeams; teamIdx++ {
		team := &Team{}
		team.skill = bestAverages[teamIdx]
		team.players = make([]*Player, teamSizes[teamIdx])
		for teammateIdx := 0; teammateIdx < teamSizes[teamIdx]; teammateIdx++ {
			team.players[teammateIdx] = &players[bestArrangement[playerIdx]]
			playerIdx++
		}
		teams.teams[teamIdx] = team
	}
	sort.Slice(teams.teams, func(i, j int) bool {
		return teams.teams[i].skill > teams.teams[j].skill
	})

	return teams
}

type Team struct {
	skill   float64
	players []*Player
}

type Teams struct {
	skillGap float64
	teams    []*Team
}

func (teams *Teams) String() string {
	longestName := 0
	for _, team := range teams.teams {
		for _, teammate := range team.players {
			if len(teammate.Name) > longestName {
				longestName = len(teammate.Name)
			}
		}
	}

	teamsStr := "```"
	for teamIdx, team := range teams.teams {
		teamsStr = fmt.Sprintf("%s\nTeam %d %s %v", teamsStr, teamIdx+1, strings.Repeat(".", longestName+1), team.skill)
		for _, teammate := range team.players {
			teamsStr = fmt.Sprintf("%s\n\t%s%s  %d", teamsStr, teammate.Name, strings.Repeat(" ", longestName-len(teammate.Name)), teammate.Skill)
		}
	}
	teamsStr = fmt.Sprintf("%s\n```", teamsStr)
	return teamsStr
}
