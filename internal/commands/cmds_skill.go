package commands

// This file handles keeping track of the skill rank of each player in the discord server

import (
	"fmt"
	"sort"

	dg "github.com/bwmarrin/discordgo"
	rsp "github.com/philflip12/spikebot/internal/responder"
	log "github.com/sirupsen/logrus"
)

func cmdSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "set":
		setSkill(session, interaction, data)
	case "increase":
		increaseSkill(session, interaction, data)
	case "decrease":
		decreaseSkill(session, interaction, data)
	case "show":
		showSkill(session, interaction, data)
	case "show_all":
		showAllSkill(session, interaction, data)
	case "guest":
		options = options[0].Options
		subCommandName := options[0].Name
		switch subCommandName {
		case "set":
			setGuestSkill(session, interaction, data)
		case "increase":
			increaseGuestSkill(session, interaction, data)
		case "decrease":
			decreaseGuestSkill(session, interaction, data)
		case "show":
			showGuestSkill(session, interaction, data)
		}
	}
}

func setSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	skill := int(options[1].IntValue())

	name, err := getUserName(data, interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	if err := data.SetPlayerSkill(userID, skill); err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Set \"%s\" skill rank to %d", name, skill)
}

func increaseSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	difference := int(options[1].IntValue())

	name, err := getUserName(data, interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	prevSkill, newSkill, err := data.ModifyPlayerSkill(userID, difference)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Increased \"%s\" skill rank from %d to %d", name, prevSkill, newSkill)
}

func decreaseSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID
	difference := int(options[1].IntValue())

	name, err := getUserName(data, interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	prevSkill, newSkill, err := data.ModifyPlayerSkill(userID, -difference)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Decreased \"%s\" skill rank from %d to %d", name, prevSkill, newSkill)
}

func showSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID

	name, err := getUserName(data, interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	player, _, err := data.GetPlayer(userID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "\"%s\" has a skill rank of %d", name, player.Skill)
}

func showAllSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
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
		str = fmt.Sprintf("%s\n%2d %s", str, player.Skill, player.Name)
	}
	str = fmt.Sprintf("%s\n```", str)

	rsp.InteractionRespond(session, interaction, str)
}
