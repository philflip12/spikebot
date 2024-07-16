package commands

// This file handles keeping track of the skill rank of each player in the discord server

import (
	"fmt"
	"sort"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func cmdSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "set":
		setSkill(session, interaction)
	case "increase":
		increaseSkill(session, interaction)
	case "decrease":
		decreaseSkill(session, interaction)
	case "show":
		showSkill(session, interaction)
	case "show_all":
		showAllSkill(session, interaction)
	}
}

func setSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	skill := int(options[1].IntValue())

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	if err := setPlayerSkill(interaction.GuildID, userID, skill); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Set \"%s\" skill rank to %d", name, skill)
}

func increaseSkill(session *dg.Session, interaction *dg.InteractionCreate) {
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

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, userID, difference)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Increased \"%s\" skill rank from %d to %d", name, prevSkill, newSkill)
}

func decreaseSkill(session *dg.Session, interaction *dg.InteractionCreate) {
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

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, userID, -difference)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Decreased \"%s\" skill rank from %d to %d", name, prevSkill, newSkill)
}

func showSkill(session *dg.Session, interaction *dg.InteractionCreate) {
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

func showAllSkill(session *dg.Session, interaction *dg.InteractionCreate) {
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
