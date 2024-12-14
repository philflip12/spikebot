package commands

// This file handles keeping track of the skill rank of each player in the discord server

import (
	"fmt"
	"sort"

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
	case "guest":
		options = options[0].Options
		subCommandName := options[0].Name
		switch subCommandName {
		case "set":
			setGuestSkill(session, interaction)
		case "increase":
			increaseGuestSkill(session, interaction)
		case "decrease":
			decreaseGuestSkill(session, interaction)
		case "show":
			showGuestSkill(session, interaction)
		}
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
	for _, player := range players {
		playerList = append(playerList, player)
	}

	sort.Slice(playerList, func(i, j int) bool {
		return playerList[i].Skill > playerList[j].Skill
	})

	str := "All Skill Ranks:\n```"
	for _, player := range playerList {
		str = fmt.Sprintf("%s\n%3d %s", str, player.Skill, player.Name)
	}
	str = fmt.Sprintf("%s\n```", str)

	if err := interactionRespond(session, interaction, str); err != nil {
		log.Error(err)
	}
}
