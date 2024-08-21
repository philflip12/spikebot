package commands

import (
	"fmt"
	"strings"

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

	// TODO: If the server contained more than 1000 people, this would need modification
	members, err := session.GuildMembers(interaction.GuildID, "", 1000)
	if err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
		return
	}

	// Crease a map of userIDs to new updated names
	nameMap := make(map[string]string, len(players))
	for _, member := range members {
		if _, ok := players[member.User.ID]; ok {
			nameMap[member.User.ID] = getNameFromMember(member)
			// remove players which were returned by query. Remaining players have left the server.
			delete(players, member.User.ID)
		}
	}

	// Save off the new names
	if err := updatePlayerNames(interaction.GuildID, nameMap); err != nil {
		log.Error(err)
		interactionRespond(session, interaction, err.Error())
	}

	responseString := "Updated the names of all tracked players"

	// Check for users who have left the server
	if len(players) > 0 {
		removeList := make([]string, 0, len(players))
		for userID := range players {
			// userID's starting with the guest prefix should not be removed.
			if !strings.HasPrefix(userID, "g") {
				removeList = append(removeList, userID)
			}
		}
		if err := deleteUsers(interaction.GuildID, removeList); err != nil {
			log.Error(err)
			interactionRespond(session, interaction, err.Error())
		}

		if len(removeList) > 0 {
			// Add information to response message about players no longer in the server
			responseString = fmt.Sprintf("%s\nRemoved players no longer in the server:", responseString)
			for _, userID := range removeList {
				player := players[userID]
				responseString = fmt.Sprintf("%s\n\t%s", responseString, player.Name)
			}
		}
	}
	interactionRespond(session, interaction, responseString)
}
