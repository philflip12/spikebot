package commands

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

const maxTeamDiffAllowed = float64(20)

func cmdTeams(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	numTeams := int(options[0].IntValue())

	players, err := getPlaying()
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

	// put an upper time limit on shuffling for good teams
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

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
		if averageDiff <= maxTeamDiffAllowed {
			break
		}
		if ctx.Err() != nil {
			break
		}
	}
	playerIndex := 0
	teamsStr := ""
	for teamIdx := 0; teamIdx < numTeams; teamIdx++ {
		averageSkill := bestAverages[teamIdx]
		teamsStr = fmt.Sprintf("%s\nTeam %d: %v", teamsStr, teamIdx+1, averageSkill)
		for teammateIdx := 0; teammateIdx < teamSizes[teamIdx]; teammateIdx++ {
			player := players[bestArrangement[playerIndex]]
			playerIndex++
			teamsStr = fmt.Sprintf("%s\n\t%s: %d", teamsStr, player.Name, player.Skill)
		}
	}

	if bestArrangementDiff <= maxTeamDiffAllowed {
		str := fmt.Sprintf("Teams found:%s", teamsStr)
		interactionRespond(session, interaction, str)
	} else {
		str := fmt.Sprintf("No valid team. Best option:%s", teamsStr)
		interactionRespond(session, interaction, str)
	}
}
