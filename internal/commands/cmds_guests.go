package commands

// This file handles keeping track of the guest players not in the discord server

import (
	"fmt"
	"sort"

	dg "github.com/bwmarrin/discordgo"
	rsp "github.com/philflip12/spikebot/internal/responder"
	log "github.com/sirupsen/logrus"
)

func cmdGuest(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options
	subCommandName := options[0].Name

	switch subCommandName {
	case "create":
		createGuest(session, interaction, data)
	case "delete":
		deleteGuest(session, interaction, data)
	case "rename":
		renameGuest(session, interaction, data)
	case "sign":
		setGuestSignature(session, interaction, data, true)
	case "unsign":
		setGuestSignature(session, interaction, data, false)
	case "show_all":
		showAllGuests(session, interaction, data)
	}
}

func createGuest(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	guestName := options[0].StringValue()
	skill := int(options[1].IntValue())
	signed := false
	if len(options) >= 3 {
		signed = options[2].BoolValue()
	}

	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}
	for _, player := range players {
		if player.Name == guestName {
			rsp.InteractionRespondf(session, interaction, "player with name %q already exists", guestName)
			return
		}
	}

	newRole := &dg.RoleParams{}
	newRole.Name = guestName + " guest"
	role, err := session.GuildRoleCreate(interaction.GuildID, newRole)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}
	guestID := "g" + role.ID

	if err := data.SaveGuest(guestID, guestName, skill, signed); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	response := fmt.Sprintf("Created guest %q", guestName)
	if skill != -1 {
		response = fmt.Sprintf("%s with skill rank %d", response, skill)
	}

	if err := data.AddPlayingUsers(guestID); err != nil {
		log.Error(err)
		errStr := fmt.Sprintf("%s\nEncountered error adding guest %q to playing group: %v", response, guestName, err)
		rsp.InteractionRespond(session, interaction, errStr)
		return
	}

	numPlayingStr, err := getNumPlayingString(data)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, "%s\nEncountered error getting playing group size", response)
		return
	}

	rsp.InteractionRespondf(session, interaction, "%s\nAdded guest %q to playing group%s", response, guestName, numPlayingStr)
}

func deleteGuest(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	if err := data.DeleteUsers(guestID); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	session.GuildRoleDelete(interaction.GuildID, roleID)

	rsp.InteractionRespondf(session, interaction, "Deleted guest %q", player.Name)
}

func renameGuest(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID
	newName := options[1].StringValue()

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	params := &dg.RoleParams{Name: newName + " guest"}
	if _, err := session.GuildRoleEdit(interaction.GuildID, roleID, params); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	if err := data.RenamePlayer(guestID, newName); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Renamed guest %q to %q", player.Name, newName)
}

func setGuestSignature(session *dg.Session, interaction *dg.InteractionCreate, data *serverData, signed bool) {
	options := interaction.ApplicationCommandData().Options[0].Options
	roleIDs := make([]string, len(options))
	for i := range options {
		roleIDs[i] = options[i].RoleValue(nil, "").ID
	}

	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
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

	err = data.UpdatePlayerSignatures(guestIDs, signed)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	response := ""
	if len(invalidRoles) != 0 {
		response = "One or more roles did not represent a guest\n\n"
	}
	action := "having signed"
	if !signed {
		action = "not having signed"
	}
	if len(guestIDs) == 1 {
		response = fmt.Sprintf("%sMarked guest %q as %s", response, players[guestIDs[0]].Name, action)
	} else if len(guestIDs) > 1 {
		response = fmt.Sprintf("%sMarked guests as %s:", response, action)
		for _, id := range guestIDs {
			response = fmt.Sprintf("%s\n\t%s", response, players[id].Name)
		}
	}
	rsp.InteractionRespond(session, interaction, response)
}

func addGuestsToPlaying(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleIDs := make([]string, len(options))
	for i := range options {
		roleIDs[i] = options[i].RoleValue(nil, "").ID
	}

	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
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

	if err := data.AddPlayingUsers(guestIDs...); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	numPlayingStr, err := getNumPlayingString(data)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	response := ""
	if len(invalidRoles) != 0 {
		response = "One or more roles did not represent a guest\n\n"
	}
	if len(guestIDs) == 1 {
		response = fmt.Sprintf("%sAdded guest %q to playing%s", response, players[guestIDs[0]].Name, numPlayingStr)
	} else if len(guestIDs) > 1 {
		response = fmt.Sprintf("%sAdded guests to playing:", response)
		for _, id := range guestIDs {
			response = fmt.Sprintf("%s\n\t%s", response, players[id].Name)
		}
		response = fmt.Sprintf("%s%s", response, numPlayingStr)
	}
	rsp.InteractionRespond(session, interaction, response)
}

func removeGuestsFromPlaying(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleIDs := make([]string, len(options))
	for i := range options {
		roleIDs[i] = options[i].RoleValue(nil, "").ID
	}

	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
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

	if err := data.RemovePlayingUsers(guestIDs...); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	numPlayingStr, err := getNumPlayingString(data)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	response := ""
	if len(invalidRoles) != 0 {
		response = "One or more roles did not represent a guest\n\n"
	}
	if len(guestIDs) == 1 {
		response = fmt.Sprintf("%sRemoved guest %q from playing%s", response, players[guestIDs[0]].Name, numPlayingStr)
	} else if len(guestIDs) > 1 {
		response = fmt.Sprintf("%sRemoved guests from playing:", response)
		for _, id := range guestIDs {
			response = fmt.Sprintf("%s\n\t%s", response, players[id].Name)
		}
		response = fmt.Sprintf("%s%s", response, numPlayingStr)
	}
	rsp.InteractionRespond(session, interaction, response)
}

func setGuestSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	skill := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	if err := data.SetPlayerSkill(guestID, skill); err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Set %q skill rank to %d", player.Name, skill)
}

func increaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	difference := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	prevSkill, newSkill, err := data.ModifyPlayerSkill(guestID, difference)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Increased guest %q skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func decreaseGuestSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	difference := int(options[1].IntValue())
	guestID := "g" + roleID

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	prevSkill, newSkill, err := data.ModifyPlayerSkill(guestID, -difference)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	rsp.InteractionRespondf(session, interaction, "Decreased guest %q skill rank from %d to %d", player.Name, prevSkill, newSkill)
}

func showGuestSkill(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	options := interaction.ApplicationCommandData().Options[0].Options[0].Options
	roleID := options[0].RoleValue(nil, "").ID
	guestID := "g" + roleID

	player, ok, err := data.GetPlayer(guestID)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}
	if !ok {
		rsp.InteractionRespondf(session, interaction, "Role selected does not represent a guest")
		return
	}

	rsp.InteractionRespondf(session, interaction, "Guest %q has a skill rank of %d", player.Name, player.Skill)
}

func showAllGuests(session *dg.Session, interaction *dg.InteractionCreate, data *serverData) {
	players, err := data.GetPlayers()
	if err != nil {
		log.Error(err)
		rsp.InteractionRespond(session, interaction, err.Error())
		return
	}

	guestList := []Player{}
	for userID, player := range players {
		if userID[0] != 'g' {
			continue
		}
		guestList = append(guestList, player)
	}

	sort.Slice(guestList, func(i, j int) bool {
		return guestList[i].Skill > guestList[j].Skill
	})

	str := "All Guests:\n```"
	for _, player := range guestList {
		str = fmt.Sprintf("%s\n%2d %s", str, player.Skill, player.Name)
	}
	str = fmt.Sprintf("%s\n```", str)

	rsp.InteractionRespond(session, interaction, str)
}
