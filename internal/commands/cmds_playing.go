package commands

// This file handles commands for tracking who is in the playing group

import (
	"fmt"
	"sort"
	"strings"

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
			addGuestToPlaying(session, interaction)
		case "remove":
			removeGuestFromPlaying(session, interaction)
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

	if len(playerList) == 0 {
		rsp.InteractionRespond(session, interaction, "The playing group is empty")
	}
	str := fmt.Sprintf("%d in playing group:\n```", len(playerList))
	for _, player := range playerList {
		str = fmt.Sprintf("%s\n%s%s  %d", str, player.Name, strings.Repeat(" ", longestName-len(player.Name)), player.Skill)
	}
	str = fmt.Sprintf("%s\n```", str)

	rsp.InteractionRespond(session, interaction, str)
}
