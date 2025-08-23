package commands

// This file handles commands for tracking who is in the playing group

import (
	"fmt"
	"sort"

	dg "github.com/bwmarrin/discordgo"
	rsp "github.com/philflip12/spikebot/internal/responder"
	log "github.com/sirupsen/logrus"
)

func cmdPlay(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "add":
		addToPlaying(session, interaction)
	case "remove":
		removeFromPlaying(session, interaction)
	case "clear":
		clearPlaying(session, interaction)
	case "show_all":
		showPlaying(session, interaction)
	case "guest":
		options = options[0].Options
		subCommandName := options[0].Name
		switch subCommandName {
		case "add":
			addGuestsToPlaying(session, interaction)
		case "remove":
			removeGuestsFromPlaying(session, interaction)
		}
	}
}

func addToPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	userIDs := make([]string, len(options))
	for i := range options {
		// Passing nil to UserValue avoids an extra API query.
		userIDs[i] = options[i].UserValue(nil).ID
	}

	names, err := getUserNames(interaction.GuildID, userIDs, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	if err := addPlayingUsers(interaction.GuildID, userIDs...); err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	numPlayingStr, err := getNumPlayingString(interaction.GuildID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	if len(names) == 1 {
		rsp.InteractionRespondf(session, interaction, "Added \"%s\" to playing%s", names[0], numPlayingStr)
		return
	}
	response := "Added users to playing:"
	for i := range names {
		response = fmt.Sprintf("%s\n\t%s", response, names[i])
	}
	response = fmt.Sprintf("%s%s", response, numPlayingStr)
	rsp.InteractionRespond(session, interaction, response)
}

func removeFromPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	userIDs := make([]string, len(options))
	for i := range options {
		// Passing nil to UserValue avoids an extra API query.
		userIDs[i] = options[i].UserValue(nil).ID
	}

	names, err := getUserNames(interaction.GuildID, userIDs, session)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	if err := removePlayingUsers(interaction.GuildID, userIDs...); err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	numPlayingStr, err := getNumPlayingString(interaction.GuildID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	if len(names) == 1 {
		rsp.InteractionRespondf(session, interaction, "Removed \"%s\" from playing%s", names[0], numPlayingStr)
		return
	}
	response := "Removed users from playing:"
	for i := range names {
		response = fmt.Sprintf("%s\n\t%s", response, names[i])
	}
	response = fmt.Sprintf("%s%s", response, numPlayingStr)
	rsp.InteractionRespond(session, interaction, response)
}

func clearPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	if err := clearPlayingUsers(interaction.GuildID); err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespond(session, interaction, "Cleared all users from playing")
}

func showPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlaying(interaction.GuildID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Skill > players[j].Skill
	})

	if len(players) == 0 {
		rsp.InteractionRespond(session, interaction, "The playing group is empty")
	}
	str := fmt.Sprintf("%d in playing group:\n```", len(players))
	for _, player := range players {
		str = fmt.Sprintf("%s\n%2d %s", str, player.Skill, player.Name)
	}
	str = fmt.Sprintf("%s\n```", str)

	rsp.InteractionRespond(session, interaction, str)
}
