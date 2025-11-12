package commands

import (
	"fmt"

	dg "github.com/bwmarrin/discordgo"
	rsp "github.com/philflip12/spikebot/internal/responder"
	log "github.com/sirupsen/logrus"
)

func cmdSign(session *dg.Session, interaction *dg.InteractionCreate, signed bool) {
	options := interaction.ApplicationCommandData().Options
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

	err = updatePlayerSignatures(interaction.GuildID, userIDs, signed)
	if err != nil {
		log.Error(err)
		rsp.InteractionRespondf(session, interaction, err.Error())
		return
	}

	action := "having signed"
	if !signed {
		action = "not having signed"
	}
	if len(names) == 1 {
		rsp.InteractionRespondf(session, interaction, "Marked \"%s\" as %s", names[0], action)
		return
	}
	response := fmt.Sprintf("Marked users as %s:", action)
	for i := range names {
		response = fmt.Sprintf("%s\n\t%s", response, names[i])
	}
	rsp.InteractionRespond(session, interaction, response)
}
