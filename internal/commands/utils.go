package commands

import (
	"errors"
	"fmt"

	dg "github.com/bwmarrin/discordgo"
)

func getUserName(guildID, userID string, session *dg.Session) (string, error) {
	name, ok, err := loadUserName(userID)
	if err != nil {
		return "", err
	}
	if !ok {
		member, err := session.GuildMember(guildID, userID)
		if err != nil {
			return "", errors.New("failed to get user info")
		}
		name = getNameFromMember(member)
		saveUserName(userID, name)
	}
	return name, nil
}

func getNameFromMember(member *dg.Member) string {
	switch {
	case member.Nick != "":
		return member.Nick
	case member.User.GlobalName != "":
		return member.User.GlobalName
	default:
		return member.User.Username
	}
}

func interactionRespondf(session *dg.Session, interaction *dg.InteractionCreate, message string, a ...any) error {
	return interactionRespond(session, interaction, fmt.Sprintf(message, a...))
}

func interactionRespond(session *dg.Session, interaction *dg.InteractionCreate, message string) error {
	return session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
		Type: dg.InteractionResponseChannelMessageWithSource,
		Data: &dg.InteractionResponseData{
			Content: message,
		},
	})
}
