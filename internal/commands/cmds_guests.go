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
	case "rename":
		renameGuest(session, interaction)
	case "show_all":
		showAllGuests(session, interaction)
	}
}

func addGuest(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	guestName := options[0].StringValue()
	skill := int(options[1].IntValue())

	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}
	for _, player := range players {
		if player.Name == guestName {
			interactionRespondf(session, interaction, "player with name %q already exists", guestName)
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
		interactionRespondf(session, interaction, "Added guest %q", guestName)
	} else {
		interactionRespondf(session, interaction, "Added guest %q with skill rank %d", guestName, skill)
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
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	if err := deleteUser(interaction.GuildID, guestID); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	session.GuildRoleDelete(interaction.GuildID, roleID)

	interactionRespondf(session, interaction, "Removed guest %q", player.Name)
}

func renameGuest(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID
	newName := options[1].StringValue()

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	params := &dg.RoleParams{Name: newName + " guest"}
	if _, err := session.GuildRoleEdit(interaction.GuildID, roleID, params); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	if err := renamePlayer(interaction.GuildID, guestID, newName); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Renamed guest %q to %q", player.Name, newName)
}

func addGuestToPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleIDs := make([]string, len(options))
	for i := range options {
		roleIDs[i] = options[i].RoleValue(nil, "").ID
	}

	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	invalidRoles := []string{}
	guestIDs := make([]string, 0, len(roleIDs))
	for i := range roleIDs {
		guestID := "g" + roleIDs[i]
		if _, ok := players[guestID]; !ok {
			invalidRoles = append(invalidRoles, roleIDs[i])
		} else {
			guestIDs = append(guestIDs, guestID)
		}
	}

	if err := addPlayingUsers(interaction.GuildID, guestIDs...); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	response := ""
	if len(invalidRoles) != 0 {
		response = "One or more roles did not represent a guest\n\n"
	}
	if len(guestIDs) == 1 {
		response = fmt.Sprintf("%sAdded guest %q to playing", response, players[guestIDs[0]].Name)
	} else if len(guestIDs) > 1 {
		response = fmt.Sprintf("%sAdded guests to playing:", response)
		for _, id := range guestIDs {
			response = fmt.Sprintf("%s\n\t%s", response, players[id].Name)
		}
	}
	interactionRespond(session, interaction, response)
}

func removeGuestFromPlaying(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleIDs := make([]string, len(options))
	for i := range options {
		roleIDs[i] = options[i].RoleValue(nil, "").ID
	}

	players, err := getPlayers(interaction.GuildID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	invalidRoles := []string{}
	guestIDs := make([]string, 0, len(roleIDs))
	for i := range roleIDs {
		guestID := "g" + roleIDs[i]
		if _, ok := players[guestID]; !ok {
			invalidRoles = append(invalidRoles, roleIDs[i])
		} else {
			guestIDs = append(guestIDs, guestID)
		}
	}

	if err := addPlayingUsers(interaction.GuildID, guestIDs...); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	response := ""
	if len(invalidRoles) != 0 {
		response = "One or more roles did not represent a guest\n\n"
	}
	if len(guestIDs) == 1 {
		response = fmt.Sprintf("%sRemoved guest %q from playing", response, players[guestIDs[0]].Name)
	} else if len(guestIDs) > 1 {
		response = fmt.Sprintf("%sRemoved guests from playing:", response)
		for _, id := range guestIDs {
			response = fmt.Sprintf("%s\n\t%s", response, players[id].Name)
		}
	}
	interactionRespond(session, interaction, response)
}

func setGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
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
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	if err := setPlayerSkill(interaction.GuildID, guestID, skill); err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Set %q skill rank to %d", player.Name, skill)
}

func increaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
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
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, guestID, difference)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Increased guest %q skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func decreaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
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
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	prevSkill, newSkill, err := modifyPlayerSkill(interaction.GuildID, guestID, -difference)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}

	interactionRespondf(session, interaction, "Decreased guest %q skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func showGuestSkill(session *dg.Session, interaction *dg.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := getPlayer(interaction.GuildID, guestID)
	if err != nil {
		log.Error(err)
		interactionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		interactionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	interactionRespondf(session, interaction, "Guest %q has a skill rank of %d", player.Name, player.Skill)
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
