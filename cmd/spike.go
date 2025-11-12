package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	dg "github.com/bwmarrin/discordgo"
	cmds "github.com/philflip12/spikebot/internal/commands"
	log "github.com/sirupsen/logrus"
)

const (
	logTimeStampFmt       = "06-01-02 15:04:05" // YYMMDD HH:MM:SS
	defaultBotTokenPath   = "./.env/BotToken"
	defaultServerIDsPath  = "./.env/ServerIDs"
	defaultChannelIDsPath = "./.env/ChannelIDs"
	defaultLogLevelStr    = "info"
	usageDialogFmtStr     = `
    SpikeBot [Options...]

    Options:
        -h                Print this help dialog
        -l LOG_LEVEL      Set program to print logs of LOG_LEVEL and higher
        -t TOKEN_PATH     Set the bot token file path to TOKEN_PATH
        -s SERVER_PATH    Set the server-id file path to SERVER_PATH
        -c CHANNEL_PATH   Set the channel-id file path to CHANNEL_PATH

    Log Levels:
        [debug, info, warn, error, fatal]

    Default Options: [SpikeBot -l "%s" -t "%s" -s "%s" -c "%s"]
`
)

type programArgs struct {
	botToken   string
	serverIDs  []string
	channelIDs []string
}

func main() {
	// Set the logrus logging formatter
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: logTimeStampFmt,
	})

	args := parseFlags()

	// Only accept commands from specified servers and channels
	cmds.SetServerIDs(args.serverIDs)
	cmds.SetChannelIDs(args.channelIDs)

	// Start spike and set the servers it will respond to commands from
	spike := startSpikeSession(args.botToken, args.serverIDs)
	defer spike.Close()

	// Register commands for the servers previously specified
	if err := spike.registerCommmands(cmds.CommandList); err != nil {
		log.Info(err)
		return
	}
	// Remove all registered commands from the servers when spike stops running.
	defer spike.deregisterCommands()

	log.Info("Spike is now running. Press CTRL-C to exit")

	waitForKillSig()
	log.Info("Closing Spike")
}

// Reads the command line flags for running the "Spike" discord bot.
// Kills the program on error or help request.
func parseFlags() *programArgs {
	var printHelp bool
	var logLevelStr string
	var botTokenPath string
	var serverIDPath string
	var channelIDPath string
	flag.BoolVar(&printHelp, "h", false, "")
	flag.StringVar(&botTokenPath, "t", defaultBotTokenPath, "h")
	flag.StringVar(&serverIDPath, "s", defaultServerIDsPath, "")
	flag.StringVar(&channelIDPath, "c", defaultChannelIDsPath, "")
	flag.StringVar(&logLevelStr, "l", defaultLogLevelStr, "")
	flag.Usage = func() {
		log.Fatalf(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath, defaultServerIDsPath, defaultChannelIDsPath)
	}
	flag.Parse()

	if printHelp {
		if printHelp {
			log.Infof(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath, defaultServerIDsPath, defaultChannelIDsPath)
		}
		os.Exit(0)
	}

	// Set log level according to provided flag or default to info
	setLogLevel(logLevelStr)

	// Open and read bot token from provided file path or default path
	botTokenFile, err := os.Open(botTokenPath)
	if err != nil {
		log.Fatalf("Failed to open bot token file path '%s'", botTokenPath)
	}
	botToken, err := io.ReadAll(botTokenFile)
	if err != nil {
		log.Fatalf("Failed to read from bot token file path '%s'", botTokenPath)
	}

	// Open and read server ID from provided file path or default path
	serverIDFile, err := os.Open(serverIDPath)
	if err != nil {
		log.Fatalf("Failed to open server ID file path '%s'", serverIDPath)
	}
	serverIDData, err := io.ReadAll(serverIDFile)
	if err != nil {
		log.Fatalf("Failed to read from server ID file path '%s'", serverIDPath)
	}
	serverReader := bytes.NewBuffer(serverIDData)
	serverScanner := bufio.NewScanner(serverReader)
	serverScanner.Split(bufio.ScanLines)
	serverIDs := []string{}
	for serverScanner.Scan() {
		serverIDs = append(serverIDs, serverScanner.Text())
	}

	// Open and read channel ID from provided file path or default path
	channelIDFile, err := os.Open(channelIDPath)
	if err != nil {
		log.Fatalf("Failed to open channel ID file path '%s'", channelIDPath)
	}
	channelIDData, err := io.ReadAll(channelIDFile)
	if err != nil {
		log.Fatalf("Failed to read from channel ID file path '%s'", channelIDPath)
	}
	channelReader := bytes.NewBuffer(channelIDData)
	channelScanner := bufio.NewScanner(channelReader)
	channelScanner.Split(bufio.ScanLines)
	channelIDs := []string{}
	for channelScanner.Scan() {
		channelIDs = append(channelIDs, channelScanner.Text())
	}

	return &programArgs{
		botToken:   string(botToken),
		serverIDs:  serverIDs,
		channelIDs: channelIDs,
	}
}

// Parses the log level flag to set the program's log level
// Kills the program on invalid log level string
func setLogLevel(levelStr string) {
	switch levelStr {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.Fatalf("Invalid log level provided '%s'", levelStr)
	}
}

type spikeSession struct {
	*dg.Session
	serverIDs []string
}

// Connects to the discord server and starts spike using the specified botToken
func startSpikeSession(botToken string, serverIDs []string) *spikeSession {
	// Configure the logging of discord go
	dg.Logger = func(msgL, caller int, format string, a ...interface{}) {
		// Start any logs from the Discord Go library with "[DG]"
		format = fmt.Sprintf("[DG] %s", format)

		switch msgL {
		case dg.LogDebug:
			log.Debugf(format, a...)
		case dg.LogInformational:
			log.Infof(format, a...)
		case dg.LogWarning:
			log.Warnf(format, a...)
		case dg.LogError:
			log.Errorf(format, a...)
		}
	}

	// Create a new Discord session using the provided bot token.
	session, err := dg.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Set the bot permission requirements ("guild" is the develement equivalent of "server")
	session.Identify.Intents = dg.IntentGuildMessages | dg.IntentGuilds | dg.IntentGuildMembers

	session.AddHandler(cmds.OnInteractionCreate)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	return &spikeSession{
		Session:   session,
		serverIDs: serverIDs,
	}

}

func (s *spikeSession) registerCommmands(cmds []*dg.ApplicationCommand) error {
	for _, serverID := range s.serverIDs {
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, cmds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *spikeSession) deregisterCommands() {
	for _, serverID := range s.serverIDs {
		s.ApplicationCommandBulkOverwrite(s.State.User.ID, serverID, nil)
	}
}

// blocks until a kill signal is sent to the program such as "CTRL-C"
func waitForKillSig() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
