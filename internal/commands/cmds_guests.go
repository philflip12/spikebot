package commands

// This file handles keeping track of the guest players not in the discord server

import (
	"fmt"
	"sort"
	"strings"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

func cmdGuest(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "add":
		addGuest(session, interaction)
	case "remove":
		removeGuest(session, interaction)
	case "playing_add":
		addGuestToPlaying(session, interaction)
	case "playing_remove":
		removeGuestFromPlaying(session, interaction)
	case "skill_set":
		setGuestSkill(session, interaction)
	case "skill_increase":
		increaseGuestSkill(session, interaction)
	case "skill_decrease":
		decreaseGuestSkill(session, interaction)
	case "skill_show":
		showGuestSkill(session, interaction)
	case "show_all":
		showAllGuests(session, interaction)
	}
}

func addGuest(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	guestName := options[0].StringValue()
	skill := -1
	if len(options) > 1 {
		skill = int(options[1].IntValue())
	}

	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}
	for _, player := range players {
		if player.Name == guestName {
			interactionRespondf(session, interaction, "player with name \"%s\" already exists", guestName)
			return
		}
	}

	newRole := &dg.RoleParams{}
	newRole.Name = guestName + " guest"
	role, err := session.GuildRoleCreate(interaction.GuildID, newRole)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}
	guestID := "g" + role.ID

	if err := saveGuest(interaction.GuildID, guestID, guestName, skill); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	if skill == -1 {
		interactionRespondf(session, interaction, "Added guest \"%s\"", guestName)
	} else {
		interactionRespondf(session, interaction, "Added guest \"%s\" with skill rank %d", guestName, skill)
	}
}

func removeGuest(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	if err := deleteGuest(interaction.GuildID, guestID); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	session.GuildRoleDelete(interaction.GuildID, roleID)

	interactionRespondf(session, interaction, "Removed guest \"%s\"", player.Name)
}

func addGuestToPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	if err := addPlayingUser(interaction.GuildID, guestID); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Added guest \"%s\" to playing", player.Name)
}

func removeGuestFromPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	if err := addPlayingUser(interaction.GuildID, roleID); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Removed guest \"%s\" to playing", player.Name)
}

func setGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	skill := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	if err := setPlayerSkill(interaction.GuildID, guestID, skill); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Set \"%s\" skill rank to %d", player.Name, skill)
}

func increaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	difference := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, guestID, difference)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Increased guest \"%s\" skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func decreaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	difference := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, guestID, -difference)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Decreased guest \"%s\" skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func showGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest.")
		return
	}

	interactionRespondf(session, interaction, "Guest \"%s\" has a skill rank of %d", player.Name, player.Skill)
}

func showAllGuests(session *dg.Session, interaction *dg.InteractionCreate) {
	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	guestList := []Player{}
	longestName := 0
	for userID, player := range players {
		if userID[0] != 'g' {
			continue
		}
		guestList = append(guestList, player)
		if len(player.Name) > longestName {
			longestName = len(player.Name)
		}
	}

	sort.Slice(guestList, func(i, j int) bool {
		return guestList[i].Skill > guestList[j].Skill
	})

	str := "All Guests:\n```"
	for _, player := range guestList {
		str = fmt.Sprintf("%s\n%s%s  %d", str, player.Name, strings.Repeat(" ", longestName-len(player.Name)), player.Skill)
	}
	str = fmt.Sprintf("%s\n```", str)

	interactionRespond(session, interaction, str)
}
