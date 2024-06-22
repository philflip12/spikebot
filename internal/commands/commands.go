package commands

import (
	"fmt"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// onMessageCreate is called every time a message is created on any channel the bot has access to.
// It is added as a callback by 'AddHandler'
func OnMessageCreate(s *dg.Session, m *dg.MessageCreate) {
	// Ignore messages created by the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	switch m.Content {
	case "@spike help":
		cmdHelp(s, m)
	case "@spike generate":
		cmdGenerate(s, m)
	case "@spike role":
		cmdRole(s, m)
	}
}

const helpMessage = "" +
	"Spike Command Options:\n" +
	"\t@spike help\n" +
	"\t@spike generate\n" +
	"\t@spike role"

func cmdHelp(s *dg.Session, m *dg.MessageCreate) {
	log.Info("Executing 'Help'")
	_, err := s.ChannelMessageSend(m.ChannelID, helpMessage)
	if err != nil {
		log.Errorf("'Generate' failed: %v", err)
	}
	log.Info("'Help' Complete")
}

func cmdGenerate(s *dg.Session, m *dg.MessageCreate) {
	log.Info("Executing 'Generate'")
	_, err := s.ChannelMessageSend(m.ChannelID, "Generated!")
	if err != nil {
		log.Errorf("'Generate' failed: %v", err)
	}
	log.Info("'Generate' Complete")
}

func cmdRole(s *dg.Session, m *dg.MessageCreate) {
	log.Info("Executing 'Role'")
	var err error
	defer func() {
		if err != nil {
			log.Errorf("'Role' failed: %v", err)
		}
		log.Info("'Role' complete")
	}()

	var roles []*dg.Role
	if roles, err = s.GuildRoles(m.GuildID); err != nil {
		return
	}

	str := "Roles:"
	for _, role := range roles {
		str += fmt.Sprintf("\n\tName: %s, ID: %s, Is Managed: %v, Is Mentionable: %v, Position: %v", role.Name, role.ID, role.Managed, role.Mentionable, role.Position)
	}

	if _, err = s.ChannelMessageSend(m.ChannelID, str); err != nil {
		return
	}
}
