package commands

// This file handles keeping track of the skill rank of each player in the discord server

import (
	"fmt"

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

	if err := setPlayerRank(userID, rank); err != nil {
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

	prevRank, newRank, err := modifyPlayerRank(userID, difference)
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

	prevRank, newRank, err := modifyPlayerRank(userID, -difference)
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

	player, _, err := getPlayer(userID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "\"%s\" has a skill rank of %d", name, player.Skill)
}

func showRanks(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlayers()
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	str := "All Skill Ranks:"
	for _, player := range players {
		str = fmt.Sprintf("%s\n  \"%s\": %d", str, player.Name, player.Skill)
	}

	interactionRespond(session, interaction, str)
}
