package commands

// This file handles commands for tracking who is in the playing group

import (
	"fmt"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func cmdPlay(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "add":
		addToPlaying(session, interaction)
	case "del":
		delFromPlaying(session, interaction)
	case "clear":
		clearPlaying(session, interaction)
	case "show":
		showPlaying(session, interaction)
	}
}

func addToPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	if err := addPlayingUser(userID); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Added \"%s\" to playing", name)
}

func delFromPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	// Passing nil to UserValue avoids an extra API query.
	userID := options[0].UserValue(nil).ID

	name, err := getUserName(interaction.GuildID, userID, session)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	if err := delPlayingUser(userID); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Removed \"%s\" from playing", name)
}

func clearPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	if err := clearPlayingUsers(); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	interactionRespond(session, interaction, "Cleared all users from playing")
}

func showPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlaying()
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	str := "Playing group:"
	for _, player := range players {
		str = fmt.Sprintf("%s\n  \"%s\": %d", str, player.Name, player.Skill)
	}

	interactionRespond(session, interaction, str)
}
