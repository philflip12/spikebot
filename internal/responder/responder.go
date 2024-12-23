package responder

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"

	dg "github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

var r ResponseManager = newResponseManager()

func InteractionRespond(session *dg.Session, interaction *dg.InteractionCreate, message string) error {
	return r.InteractionRespond(session, interaction, message)
}

func InteractionRespondf(session *dg.Session, interaction *dg.InteractionCreate, message string, a ...any) error {
	return r.InteractionRespondf(session, interaction, message, a...)
}

func InteractionContinue(session *dg.Session, interaction *dg.InteractionCreate) error {
	return r.InteractionContinue(session, interaction)
}

type ResponseManager interface {
	InteractionRespond(session *dg.Session, interaction *dg.InteractionCreate, message string) error
	InteractionRespondf(session *dg.Session, interaction *dg.InteractionCreate, message string, a ...any) error
	InteractionContinue(session *dg.Session, interaction *dg.InteractionCreate) error
}

type responseBuffer struct {
	message string
	mutex   sync.Mutex
}

type responseManager struct {
	responseBuffers map[string]*responseBuffer
	mapLock         sync.RWMutex
}

func newResponseManager() *responseManager {
	return &responseManager{
		responseBuffers: map[string]*responseBuffer{},
	}
}

const maxMessageLen = 2000
const maxSplitCutoff = 100

const continuePrompt = "\n...\n*/continue to show more*"

var continuePromptLen = len(continuePrompt)
var maxSplitIndex = maxMessageLen - continuePromptLen
var minSplitIndex = maxSplitIndex - maxSplitCutoff

func (r *responseManager) InteractionRespond(
	session *dg.Session,
	interaction *dg.InteractionCreate,
	message string,
) error {
	if len(message) <= maxMessageLen {
		err := session.InteractionRespond(interaction.Interaction, &dg.InteractionResponse{
			Type: dg.InteractionResponseChannelMessageWithSource,
			Data: &dg.InteractionResponseData{
				Content: message,
			},
		})
		if err != nil {
			log.Error(err.Error())
		}
		return err
	}
	splitIndex := maxSplitIndex
	if index := strings.LastIndex(message[minSplitIndex:maxSplitIndex+1], "\n"); index != -1 {
		splitIndex = minSplitIndex + index
	} else if index := strings.LastIndexFunc(message[minSplitIndex:maxSplitIndex+1], func(r rune) bool {
		return unicode.IsSpace(r)
	}); index != -1 {
		splitIndex = minSplitIndex + index
	}
	messageBuff := make([]byte, 0, splitIndex+continuePromptLen)
	messageBuff = append(messageBuff, message[:splitIndex]...)
	messageBuff = append(messageBuff, continuePrompt...)
	err := r.InteractionRespond(session, interaction, string(messageBuff))
	if err == nil {
		r.updateBuffer(interaction.GuildID, strings.TrimSpace(string([]byte(message)[splitIndex:])))
	}
	return err
}

func (r *responseManager) InteractionRespondf(
	session *dg.Session,
	interaction *dg.InteractionCreate,
	message string,
	a ...any,
) error {
	return r.InteractionRespond(session, interaction, fmt.Sprintf(message, a...))
}

var ErrNoResponseContinuation = errors.New("no response output to continue")

func (r *responseManager) InteractionContinue(
	session *dg.Session,
	interaction *dg.InteractionCreate,
) error {
	message := r.getBuffer(interaction.GuildID)
	if message == "" {
		return ErrNoResponseContinuation
	}
	return r.InteractionRespond(session, interaction, message)
}

func (r *responseManager) updateBuffer(guildID string, message string) {
	r.mapLock.RLock()
	buffer, ok := r.responseBuffers[guildID]
	r.mapLock.RUnlock()
	if ok {
		buffer.mutex.Lock()
		buffer.message = message
		buffer.mutex.Unlock()
		return
	}

	r.mapLock.Lock()
	buffer, ok = r.responseBuffers[guildID]
	if !ok {
		buffer = &responseBuffer{
			message: message,
		}
		r.responseBuffers[guildID] = buffer
		r.mapLock.Unlock()
		return
	}
	r.mapLock.Unlock()

	buffer.mutex.Lock()
	buffer.message = message
	buffer.mutex.Unlock()
}

func (r *responseManager) getBuffer(guildID string) string {
	r.mapLock.RLock()
	buffer, ok := r.responseBuffers[guildID]
	r.mapLock.RUnlock()
	if !ok {
		return ""
	}
	buffer.mutex.Lock()
	message := buffer.message
	buffer.message = ""
	buffer.mutex.Unlock()
	return message
}
