package commands

import (
	"errors"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func cmdLastActive(session *dg.Session, interaction *dg.InteractionCreate) {
	interactionRespond(session, interaction, "Not yet implemented")
}

func cmdUpdateNames(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	nameMap := make(map[string]string, len(players))

	for userID := range players {
		member, err := session.GuildMember(interaction.GuildID, userID)
		if err != nil {
			err = errors.New("failed to get user info")
			log.Error(err)
			interactionRespond(session, interaction, err.Error())
		}
		nameMap[userID] = getNameFromMember(member)
	}

	if err := updatePlayerNames(interaction.GuildID, nameMap); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
	}
	interactionRespond(session, interaction, "Updated the names of all tracked players")
}
