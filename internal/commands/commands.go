package commands

import (
	"fmt"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// onMessageCreate is called every time a message is created on any channel the bot has access to.
// It is added as a callback by 'AddHandler'
func OnInteractionCreate(s *dg.Session, i *dg.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "help":
		cmdHelp(s, i)
	case "generate":
		cmdGenerate(s, i)
	case "role":
		cmdRole(s, i)
	}
}

// func OnInteractionEvent(s *dg.Session, m *dg.InteractionCreate)

var CommandList = []*dg.ApplicationCommand{
	{
		Name:        "help",
		Description: "Print all spike commands",
	},
	{
		Name:        "generate",
		Description: "Not Yet Implemented",
	},
	{
		Name:        "role",
		Description: "Not Yet Implemented",
	},
}

const helpMessage = "" +
	"Spike Command Options:\n" +
	"\t/help\n" +
	"\t/generate\n" +
	"\t/role"

func cmdHelp(s *dg.Session, i *dg.InteractionCreate) {
	log.Info("Executing 'Help'")
	if err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Content: helpMessage,
		},
	}); err != nil {
		log.Errorf("'Help' failed: %v", err)
		return
	}
	log.Info("'Help' Complete")
}

func cmdGenerate(s *dg.Session, i *dg.InteractionCreate) {
	log.Info("Executing 'Generate'")
	if err := s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Content: "Generated!",
		},
	}); err != nil {
		log.Errorf("'Generate' failed: %v", err)
		return
	}
	log.Info("'Generate' Complete")
}

func cmdRole(s *dg.Session, i *dg.InteractionCreate) {
	log.Info("Executing 'Role'")
	var err error
	defer func() {
		if err != nil {
			log.Errorf("'Role' failed: %v", err)
		} else {
			log.Info("'Role' complete")
		}
	}()

	var roles []*dg.Role
	if roles, err = s.GuildRoles(i.GuildID); err != nil {
		return
	}

	str := "Roles:"
	for _, role := range roles {
		str += fmt.Sprintf("\n\tName: %s, ID: %s, Is Managed: %v, Is Mentionable: %v, Position: %v", role.Name, role.ID, role.Managed, role.Mentionable, role.Position)
	}

	err = s.InteractionRespond(i.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Content: str,
		},
	})
}
