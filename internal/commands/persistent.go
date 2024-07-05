package commands

// This file contains functionality for the storage and retrieval of persistent data which should be
// saved between commands and bot restarts.

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type LoadSaver interface {
	Lock()
	Unlock()
	Load() error
	Save() error
}

const persistentDataDirectory = "persistentData/"
const playerDataFileName = "playerData"
const playingListFileName = "playingList"

var players persistentObject[map[string]Player]
var playing persistentObject[map[string]struct{}]

func init() {
	players.fileName = playerDataFileName
	players.makeNew = func() map[string]Player { return map[string]Player{} }
	playing.fileName = playingListFileName
	playing.makeNew = func() map[string]struct{} { return map[string]struct{}{} }
}

type persistentObject[T any] struct {
	Mutex    sync.Mutex
	isLoaded bool
	fileName string
	object   T
	makeNew  func() T
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

	if _, err := os.Stat(persistentDataDirectory); errors.Is(err, os.ErrNotExist) {
		p.object = p.makeNew()
		p.isLoaded = true
		return nil
	}

	if _, err := os.Stat(persistentDataDirectory + p.fileName + ".json"); errors.Is(err, os.ErrNotExist) {
		p.object = p.makeNew()
		p.isLoaded = true
		return nil
	}

	data, err := os.ReadFile(persistentDataDirectory + p.fileName + ".json")
	if err != nil {
		return err
	}

	var object T
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}

	p.object = object
	p.isLoaded = true
	return nil
}

func (p *persistentObject[T]) Save() error {
	data, err := json.Marshal(p.object)
	if err != nil {
		return err
	}

	if _, err := os.Stat(persistentDataDirectory); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(persistentDataDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	if err := os.WriteFile(persistentDataDirectory+p.fileName+"_temp.json", data, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(persistentDataDirectory+p.fileName+"_temp.json", persistentDataDirectory+p.fileName+".json")
}

type Player struct {
	Name  string `json:"name"`
	Skill int    `json:"skill"`
}

func loadUserName(userID string) (string, bool, error) {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return "", false, err
	}

	player, ok := players.object[userID]
	if !ok {
		return "", false, nil
	}
	return player.Name, true, nil
}

func saveUserName(userID string, name string) error {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return err
	}

	if _, ok := players.object[userID]; ok {
		return nil
	}
	players.object[userID] = Player{
		Name:  name,
		Skill: -1,
	}
	return players.Save()
}

func addPlayingUser(userID string) error {
	playing.Lock()
	defer playing.Unlock()
	if err := playing.Load(); err != nil {
		return err
	}

	if _, ok := playing.object[userID]; !ok {
		playing.object[userID] = struct{}{}
		return playing.Save()
	}
	return nil
}

func delPlayingUser(userID string) error {
	playing.Lock()
	defer playing.Unlock()
	if err := playing.Load(); err != nil {
		return err
	}

	if _, ok := playing.object[userID]; ok {
		delete(playing.object, userID)
		return playing.Save()
	}
	return nil
}

func clearPlayingUsers() error {
	playing.Lock()
	defer playing.Unlock()

	playing.object = map[string]struct{}{}
	return playing.Save()
}

func getPlaying() ([]Player, error) {
	playing.Lock()
	if err := playing.Load(); err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(playing.object))
	for key := range playing.object {
		userIDs = append(userIDs, key)
	}
	playing.Unlock()

	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return nil, err
	}

	result := make([]Player, 0, len(userIDs))
	for _, userID := range userIDs {
		result = append(result, players.object[userID])
	}

	return result, nil
}

func setPlayerRank(userID string, rank int) error {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return err
	}

	player, ok := players.object[userID]
	if !ok {
		return nil
	}

	player.Skill = rank
	players.object[userID] = player
	return players.Save()
}

func modifyPlayerRank(userID string, diff int) (prev, new int, err error) {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return 0, 0, err
	}

	if player, ok := players.object[userID]; ok {
		prev = player.Skill
		player.Skill += diff
		if player.Skill > 100 {
			player.Skill = 100
		} else if player.Skill < 0 {
			player.Skill = 0
		}
		new = player.Skill
		players.object[userID] = player
	}
	return prev, new, players.Save()
}

func getPlayer(userID string) (Player, bool, error) {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return Player{}, false, err
	}

	player, ok := players.object[userID]
	return player, ok, nil
}

func getPlayers() ([]Player, error) {
	players.Lock()
	defer players.Unlock()
	if err := players.Load(); err != nil {
		return nil, err
	}

	allPlayers := make([]Player, 0, len(players.object))
	for userID := range players.object {
		allPlayers = append(allPlayers, players.object[userID])
	}
	return allPlayers, nil
}
