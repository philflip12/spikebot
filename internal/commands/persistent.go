package commands

// This file contains functionality for the storage and retrieval of persistent data which should be
// saved between commands and bot restarts.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

const persistentDataDirectory = "persistentData"
const playerDataFileName = "playerData"
const playingListFileName = "playingList"

// map[serverID]persistentObject[map[userID]Player]
var players map[string]*persistentObject[map[string]Player]

// map[serverID]persistentObject[map[userID]exists]
var playing map[string]*persistentObject[map[string]struct{}]

// initializes the players and playing persistentObject variables for each server being serviced
func setPlayersAndPlayingServerIDs(serverIDs []string) {
	players = make(map[string]*persistentObject[map[string]Player], len(serverIDs))
	playing = make(map[string]*persistentObject[map[string]struct{}], len(serverIDs))
	for _, serverID := range serverIDs {
		players[serverID] = &persistentObject[map[string]Player]{
			filePath:   fmt.Sprintf("%s/%s", persistentDataDirectory, serverID),
			fileName:   playerDataFileName,
			makeNew:    func() map[string]Player { return map[string]Player{} },
			checkValid: func(m map[string]Player) bool { return m != nil },
		}
		playing[serverID] = &persistentObject[map[string]struct{}]{
			filePath:   fmt.Sprintf("%s/%s", persistentDataDirectory, serverID),
			fileName:   playingListFileName,
			makeNew:    func() map[string]struct{} { return map[string]struct{}{} },
			checkValid: func(m map[string]struct{}) bool { return m != nil },
		}
	}
}

type persistentObject[T any] struct {
	Mutex      sync.Mutex
	isLoaded   bool
	filePath   string
	fileName   string
	object     T
	makeNew    func() T
	checkValid func(T) bool
}

func (p *persistentObject[T]) Lock() {
	p.Mutex.Lock()
}

func (p *persistentObject[T]) Unlock() {
	p.Mutex.Unlock()
}

func (p *persistentObject[T]) Load() error {
	if p.isLoaded {
		return nil
	}

	if _, err := os.Stat(p.filePath); errors.Is(err, os.ErrNotExist) {
		p.object = p.makeNew()
		p.isLoaded = true
		return nil
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s.json", p.filePath, p.fileName)); errors.Is(err, os.ErrNotExist) {
		p.object = p.makeNew()
		p.isLoaded = true
		return nil
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s.json", p.filePath, p.fileName))
	if err != nil {
		return err
	}

	var object T
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	if !p.checkValid(object) {
		object = p.makeNew()
	}

	p.object = object
	p.isLoaded = true
	return nil
}

func (p *persistentObject[T]) Save() error {
	if !p.checkValid(p.object) {
		return errors.New("saved object is not valid")
	}
	data, err := json.Marshal(p.object)
	if err != nil {
		return err
	}

	if _, err := os.Stat(p.filePath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(p.filePath, os.ModePerm); err != nil {
			return err
		}
	}

	if err := os.WriteFile(fmt.Sprintf("%s/%s_temp.json", p.filePath, p.fileName), data, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(fmt.Sprintf("%s/%s_temp.json", p.filePath, p.fileName), fmt.Sprintf("%s/%s.json", p.filePath, p.fileName))
}

type Player struct {
	Name  string `json:"name"`
	Skill int    `json:"skill"`
}

func loadUserName(serverID string, userID string) (string, bool, error) {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return "", false, err
	}

	player, ok := players[serverID].object[userID]
	if !ok {
		return "", false, nil
	}
	return player.Name, true, nil
}

func saveUserName(serverID string, userID string, name string) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	if _, ok := players[serverID].object[userID]; ok {
		return nil
	}
	players[serverID].object[userID] = Player{
		Name:  name,
		Skill: -1,
	}
	return players[serverID].Save()
}

func deleteUser(serverID, userID string) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	if _, ok := players[serverID].object[userID]; !ok {
		return errors.New("cannot delete user: ID not found")
	}

	delete(players[serverID].object, userID)
	if err := players[serverID].Save(); err != nil {
		return err
	}

	// ensure the user is deleted from the playing group as well if they are in it.
	playing[serverID].Lock()
	defer playing[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	if _, ok := playing[serverID].object[userID]; !ok {
		return nil
	}

	delete(playing[serverID].object, userID)
	return playing[serverID].Save()
}

func deleteUsers(serverID string, userIDs []string) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if _, ok := players[serverID].object[userID]; !ok {
			return errors.New("cannot delete user: ID not found")
		}

		delete(players[serverID].object, userID)
		if err := players[serverID].Save(); err != nil {
			return err
		}
	}

	// ensure the guest is deleted from the playing group as well if they are in it.
	playing[serverID].Lock()
	defer playing[serverID].Unlock()
	if err := playing[serverID].Load(); err != nil {
		return err
	}

	for _, userID := range userIDs {
		if _, ok := playing[serverID].object[userID]; !ok {
			continue
		}

		delete(playing[serverID].object, userID)
	}

	return playing[serverID].Save()
}

func addPlayingUsers(serverID string, userIDs ...string) error {
	playing[serverID].Lock()
	defer playing[serverID].Unlock()
	if err := playing[serverID].Load(); err != nil {
		return err
	}

	change := false
	for i := range userIDs {
		if _, ok := playing[serverID].object[userIDs[i]]; !ok {
			playing[serverID].object[userIDs[i]] = struct{}{}
			change = true
		}
	}
	if change {
		return playing[serverID].Save()
	}
	return nil
}

func removePlayingUsers(serverID string, userIDs ...string) error {
	playing[serverID].Lock()
	defer playing[serverID].Unlock()
	if err := playing[serverID].Load(); err != nil {
		return err
	}

	change := false
	for i := range userIDs {
		if _, ok := playing[serverID].object[userIDs[i]]; ok {
			delete(playing[serverID].object, userIDs[i])
			change = true
		}
	}
	if change {
		return playing[serverID].Save()
	}
	return nil
}

func clearPlayingUsers(serverID string) error {
	playing[serverID].Lock()
	defer playing[serverID].Unlock()

	playing[serverID].object = map[string]struct{}{}
	return playing[serverID].Save()
}

func getPlaying(serverID string) ([]Player, error) {
	playing[serverID].Lock()
	if err := playing[serverID].Load(); err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(playing[serverID].object))
	for key := range playing[serverID].object {
		userIDs = append(userIDs, key)
	}
	playing[serverID].Unlock()

	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return nil, err
	}

	result := make([]Player, 0, len(userIDs))
	for _, userID := range userIDs {
		result = append(result, players[serverID].object[userID])
	}

	return result, nil
}

func getPlayingCount(serverID string) (int, error) {
	playing[serverID].Lock()
	defer playing[serverID].Unlock()
	if err := playing[serverID].Load(); err != nil {
		return 0, err
	}

	return len(playing[serverID].object), nil
}

func setPlayerSkill(serverID string, userID string, skill int) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	player, ok := players[serverID].object[userID]
	if !ok {
		return nil
	}

	player.Skill = skill
	players[serverID].object[userID] = player
	return players[serverID].Save()
}

func modifyPlayerSkill(serverID string, userID string, diff int) (prev, new int, err error) {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return 0, 0, err
	}

	if player, ok := players[serverID].object[userID]; ok {
		prev = player.Skill
		player.Skill += diff
		if player.Skill > 100 {
			player.Skill = 100
		} else if player.Skill < 0 {
			player.Skill = 0
		}
		new = player.Skill
		players[serverID].object[userID] = player
	}
	return prev, new, players[serverID].Save()
}

func getPlayer(serverID string, userID string) (Player, bool, error) {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return Player{}, false, err
	}

	player, ok := players[serverID].object[userID]
	return player, ok, nil
}

func getPlayers(serverID string) (map[string]Player, error) {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return nil, err
	}

	playerMap := make(map[string]Player, len(players[serverID].object))
	for userID, player := range players[serverID].object {
		playerMap[userID] = player
	}
	return playerMap, nil
}

func updatePlayerNames(serverID string, nameMap map[string]string) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	for userID, name := range nameMap {
		player, ok := players[serverID].object[userID]
		if !ok {
			continue
		}
		player.Name = name
		players[serverID].object[userID] = player
	}

	return players[serverID].Save()
}

func saveGuest(serverID, guestID, guestName string, skill int) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	if _, ok := players[serverID].object[guestID]; ok {
		return fmt.Errorf("cannot save guest \"%s\": ID already in use", guestName)
	}

	players[serverID].object[guestID] = Player{
		Name:  guestName,
		Skill: skill,
	}

	return players[serverID].Save()
}

func renamePlayer(serverID, guestID, guestName string) error {
	players[serverID].Lock()
	defer players[serverID].Unlock()
	if err := players[serverID].Load(); err != nil {
		return err
	}

	player, ok := players[serverID].object[guestID]
	if !ok {
		return fmt.Errorf("guest with id %s not found", guestID)
	}

	player.Name = guestName
	players[serverID].object[guestID] = player

	return players[serverID].Save()
}
