package commands

import dg "github.com/bwmarrin/discordgo"

func cmdLastActive(session *dg.Session, interaction *dg.InteractionCreate) {
	interactionRespond(session, interaction, "Not yet implemented")
}
