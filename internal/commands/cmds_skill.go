package commands

// This file handles keeping track of the skill rank of each player in the discord server

import (
	"fmt"
	"sort"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func cmdRank(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "set":
		setRank(session, interaction)
	case "increase":
		increaseRank(session, interaction)
	case "decrease":
		decreaseRank(session, interaction)
	case "show":
		showRank(session, interaction)
	case "show_all":
		showRanks(session, interaction)
	}
}

func setRank(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	rank := int(options[1].IntValue())

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	if err := setPlayerRank(interaction.GuildID, userID, rank); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Set \"%s\" skill rank to %d", name, rank)
}

func increaseRank(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	difference := int(options[1].IntValue())

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	prevRank, newRank, err := modifyPlayerRank(interaction.GuildID, userID, difference)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Increased \"%s\" skill rank from %d to %d", name, prevRank, newRank)
}

func decreaseRank(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	difference := int(options[1].IntValue())

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	prevRank, newRank, err := modifyPlayerRank(interaction.GuildID, userID, -difference)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Decreased \"%s\" skill rank from %d to %d", name, prevRank, newRank)
}

func showRank(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	player, _, err := getPlayer(interaction.GuildID, userID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "\"%s\" has a skill rank of %d", name, player.Skill)
}

func showRanks(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	playerList := make([]Player, 0, len(players))
	longestName := 0
	for _, player := range players {
		playerList = append(playerList, player)
		if len(player.Name) > longestName {
			longestName = len(player.Name)
		}
	}

	sort.Slice(playerList, func(i, j int) bool {
		return playerList[i].Skill > playerList[j].Skill
	})

	str := "All Skill Ranks:\n```"
	for _, player := range playerList {
		str = fmt.Sprintf("%s\n%s%s  %d", str, player.Name, strings.Repeat(" ", longestName-len(player.Name)), player.Skill)
	}
	str = fmt.Sprintf("%s\n```", str)

	interactionRespond(session, interaction, str)
}
