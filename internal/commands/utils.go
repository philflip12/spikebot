package commands

import (
	"errors"
	"fmt"

	dg "github.com/bwmarrin/discordgo"
)

func getUserName(serverID, userID string, session *dg.Session) (string, error) {
	name, ok, err := loadUserName(serverID, userID)
	if err != nil {
		return "", err
	}
	if !ok {
		member, err := session.GuildMember(serverID, userID)
		if err != nil {
			return "", errors.New("failed to get user info")
		}
		name = getNameFromMember(member)
		saveUserName(serverID, userID, name)
	}
	return name, nil
}

func getUserNames(serverID string, userIDs []string, session *dg.Session) ([]string, error) {
	players, err := getPlayers(serverID)
	if err != nil {
		return nil, fmt.Errorf("error loading players: %w", err)
	}

	names := make([]string, 0, len(userIDs))
	missingIDs := map[string]struct{}{}
	for _, id := range userIDs {
		player, ok := players[id]
		if ok {
			names = append(names, player.Name)
		} else {
			missingIDs[id] = struct{}{}
		}
	}

	switch len(missingIDs) {
	case 0:
	case 1:
		for id := range missingIDs {
			member, err := session.GuildMember(serverID, id)
			if err != nil {
				return nil, fmt.Errorf("failed to get user info: %w", err)
			}
			name := getNameFromMember(member)
			names = append(names, name)
			saveUserName(serverID, id, name)
		}
	default:
		members, err := session.GuildMembers(serverID, "", 1000)
		if err != nil {
			return nil, fmt.Errorf("failed to get users' info: %w", err)
		}
		missingRemaining := len(missingIDs)
		for _, member := range members {
			if _, ok := missingIDs[member.User.ID]; ok {
				name := getNameFromMember(member)
				names = append(names, name)
				saveUserName(serverID, member.User.ID, name)
				missingRemaining--
				if missingRemaining == 0 {
					break
				}
			}
		}
	}
	return names, nil
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

func getNumPlayingString(serverID string) (string, error) {
	numPlaying, err := getPlayingCount(serverID)
	if err != nil {
		return "", err
	}
	if numPlaying == 0 {
		return "", nil
	}
	return fmt.Sprintf("\n%d in playing group", numPlaying), nil
}

func SetServerIDs(serverIDs []string) {
	setPlayersAndPlayingServerIDs(serverIDs)
	setLastTeamsOptionsServerIDs(serverIDs)
}
