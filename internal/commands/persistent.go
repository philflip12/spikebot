package commands

// This file contains functionality for the storage and retrieval of persistent data which should be
// saved between commands and bot restarts.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/philflip12/spikebot/pkg/atomic"
)

const (
	persistentDataDirectory = "persistentData"
	settingsFileName        = "settings"
	playerDataFileName      = "playerData"
	playingListFileName     = "playingList"
)

var servers = atomic.NewAtomicMap[string, *serverData]()

func newServerData(serverID string) *serverData {
	filepath.Join(persistentDataDirectory, serverID)
	serverDirectory := fmt.Sprintf("%s/%s", persistentDataDirectory, serverID)
	return &serverData{
		Settings: persistentObject[*Settings]{
			filePath:   serverDirectory,
			fileName:   settingsFileName,
			makeNew:    func() *Settings { return &Settings{} },
			checkValid: func(m *Settings) bool { return m != nil },
		},
		Players: persistentObject[map[string]Player]{
			filePath:   serverDirectory,
			fileName:   playerDataFileName,
			makeNew:    func() map[string]Player { return map[string]Player{} },
			checkValid: func(m map[string]Player) bool { return m != nil },
		},
		Playing: persistentObject[map[string]struct{}]{
			filePath:   serverDirectory,
			fileName:   playingListFileName,
			makeNew:    func() map[string]struct{} { return map[string]struct{}{} },
			checkValid: func(m map[string]struct{}) bool { return m != nil },
		},
	}
}

type serverData struct {
	Settings persistentObject[*Settings]
	Players  persistentObject[map[string]Player]
	Playing  persistentObject[map[string]struct{}]
}

type Settings struct {
	RequireSignatures bool `json:"requireSignatures"`
}

// initializes the players and playing persistentObject variables for each server being serviced
func setPlayersAndPlayingServerIDs(serverIDs []string) {
	for _, serverID := range serverIDs {
		servers.Write(serverID, newServerData(serverID))
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

func (p *persistentObject[T]) WithLock(do func(object T) (dirty bool)) error {
	p.Lock()
	defer p.Unlock()
	if err := p.Load(); err != nil {
		return err
	}

	dirty := do(p.object)

	if dirty {
		return p.Save()
	}
	return nil
}

func (d *serverData) SetSignatureRequirement(isRequired bool) error {
	return d.Settings.WithLock(func(s *Settings) (dirty bool) {
		wasRequired := s.RequireSignatures
		s.RequireSignatures = isRequired
		return wasRequired != isRequired
	})
}

func (d *serverData) GetSettings() (Settings, error) {
	var settings Settings
	err := d.Settings.WithLock(func(s *Settings) (dirty bool) {
		settings = *s
		return false
	})
	return settings, err
}

type Player struct {
	Name   string `json:"name"`
	Skill  int    `json:"skill"`
	Signed bool   `json:"signed"`
}

func (d *serverData) LoadUserName(userID string) (string, bool, error) {
	name, ok := "", false
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		player, ok := players[userID]
		if ok {
			name = player.Name
		}
		return false
	})
	return name, ok, err
}

func (d *serverData) SaveUserName(userID string, name string) error {
	return d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		if _, ok := players[userID]; ok {
			return false
		}
		players[userID] = Player{
			Name:   name,
			Skill:  -1,
			Signed: false,
		}
		return true
	})
}

func (d *serverData) DeleteUsers(userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Remove from playing group before deleting from database
	err := d.Playing.WithLock(func(playing map[string]struct{}) (dirty bool) {
		for _, userID := range userIDs {
			if _, ok := playing[userID]; ok {
				delete(playing, userID)
				dirty = true
			}
		}
		return dirty
	})
	if err != nil {
		return err
	}

	missingIDs := 0
	err = d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		for _, userID := range userIDs {
			if _, ok := players[userID]; !ok {
				missingIDs++
			} else {
				delete(players, userID)
				dirty = true
			}
		}
		return dirty
	})
	if err != nil {
		return err
	}
	switch missingIDs {
	case 0:
		return nil
	case 1:
		return errors.New("could not delete user: ID not found")
	default:
		return fmt.Errorf("could not delete %d users: ID not found", missingIDs)
	}
}

func (d *serverData) AddPlayingUsers(userIDs ...string) error {
	return d.Playing.WithLock(func(playing map[string]struct{}) (dirty bool) {
		for i := range userIDs {
			if _, ok := playing[userIDs[i]]; !ok {
				playing[userIDs[i]] = struct{}{}
				dirty = true
			}
		}
		return dirty
	})
}

func (d *serverData) RemovePlayingUsers(userIDs ...string) error {
	return d.Playing.WithLock(func(playing map[string]struct{}) (dirty bool) {
		for i := range userIDs {
			if _, ok := playing[userIDs[i]]; ok {
				delete(playing, userIDs[i])
				dirty = true
			}
		}
		return dirty
	})
}

func (d *serverData) ClearPlayingUsers() error {
	d.Playing.Lock()
	defer d.Playing.Unlock()

	d.Playing.object = map[string]struct{}{}
	return d.Playing.Save()
}

func (d *serverData) GetPlaying() ([]Player, error) {
	var userIDs []string
	err := d.Playing.WithLock(func(playing map[string]struct{}) (dirty bool) {
		userIDs = make([]string, 0, len(playing))
		for key := range playing {
			userIDs = append(userIDs, key)
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	playingPlayers := make([]Player, 0, len(userIDs))
	var mapErr error
	err = d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		for _, userID := range userIDs {
			player, ok := players[userID]
			if !ok {
				mapErr = errors.New("playing userID not found in list of players")
				return false
			}
			playingPlayers = append(playingPlayers, player)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if mapErr != nil {
		return nil, mapErr
	}
	return playingPlayers, nil
}

func (d *serverData) GetPlayingCount() (int, error) {
	var count int
	err := d.Playing.WithLock(func(playing map[string]struct{}) (dirty bool) {
		count = len(playing)
		return false
	})
	return count, err
}

func (d *serverData) SetPlayerSkill(userID string, skill int) error {
	var mapErr error
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		player, ok := players[userID]
		if !ok {
			mapErr = errors.New("userID not found in list of players")
			return false
		}
		skillBefore := player.Skill
		player.Skill = skill
		players[userID] = player
		return skill != skillBefore
	})
	if err != nil {
		return err
	}
	return mapErr
}

func (d *serverData) ModifyPlayerSkill(userID string, diff int) (prev, new int, err error) {
	var mapErr error
	err = d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		player, ok := players[userID]
		if !ok {
			mapErr = errors.New("userID not found in list of players")
			return false
		}
		prev = player.Skill
		player.Skill += diff
		if player.Skill > 99 {
			player.Skill = 99
		} else if player.Skill < 0 {
			player.Skill = 0
		}
		new = player.Skill
		players[userID] = player
		return prev != new
	})
	if err != nil {
		return 0, 0, err
	}
	if mapErr != nil {
		return 0, 0, mapErr
	}
	return prev, new, nil
}

func (d *serverData) UpdatePlayerSignatures(userIDs []string, signed bool) error {
	missingIDs := 0
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		for _, userID := range userIDs {
			player, ok := players[userID]
			if !ok {
				missingIDs++
				continue
			}
			player.Signed = signed
			players[userID] = player
			dirty = true
		}
		return dirty
	})
	if err != nil {
		return err
	}
	switch missingIDs {
	case 0:
		return nil
	case 1:
		return errors.New("could modify user: ID not found")
	default:
		return fmt.Errorf("could not modify %d users: ID not found", missingIDs)
	}
}

func (d *serverData) GetPlayer(userID string) (Player, bool, error) {
	var player Player
	var found bool
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		player, found = players[userID]
		return false
	})
	if err != nil {
		return Player{}, false, err
	}
	return player, found, nil
}

func (d *serverData) GetPlayers() (map[string]Player, error) {
	var playerMap map[string]Player
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		playerMap = make(map[string]Player, len(players))
		for userID, player := range players {
			playerMap[userID] = player
		}
		return false
	})
	return playerMap, err
}

func (d *serverData) UpdatePlayerNames(nameMap map[string]string) error {
	return d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		dirty = false
		for userID, name := range nameMap {
			player, ok := players[userID]
			if !ok {
				continue
			}
			if player.Name != name {
				dirty = true
			}
			player.Name = name
			players[userID] = player
		}
		return dirty
	})
}

func (d *serverData) SaveGuest(guestID, guestName string, skill int, signed bool) error {
	var mapErr error
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		if _, ok := players[guestID]; ok {
			mapErr = fmt.Errorf("cannot save guest \"%s\": ID already in use", guestName)
			return false
		}
		players[guestID] = Player{
			Name:   guestName,
			Skill:  skill,
			Signed: signed,
		}
		return true
	})
	if err != nil {
		return err
	}
	return mapErr
}

func (d *serverData) RenamePlayer(guestID, guestName string) error {
	var mapErr error
	err := d.Players.WithLock(func(players map[string]Player) (dirty bool) {
		player, ok := players[guestID]
		if !ok {
			mapErr = fmt.Errorf("guest with id %s not found", guestID)
			return false
		}
		nameBefore := player.Name
		player.Name = guestName
		players[guestID] = player
		return guestName != nameBefore
	})
	if err != nil {
		return err
	}
	return mapErr
}
