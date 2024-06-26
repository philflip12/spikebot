package main

import (
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"

	dg "github.com/bwmarrin/discordgo"
	cmds "github.com/philflip12/spikebot/internal/commands"
	log "github.com/sirupsen/logrus"
)

const (
	logTimeStampFmt     = "06-01-02 3:04:05 pm" // YY-MM-DD HH:MM:SS pm
	defaultBotTokenPath = "./.env/BotToken"
	defaultGuildIDPath  = "./.env/ServerID"
	defaultLogLevelStr  = "info"
	usageDialogFmtStr   = `
    SpikeBot [Options...]

    Options:
        -h               Print this help dialog
        -l LOG_LEVEL     Set program to print logs of LOG_LEVEL and higher
        -t TOKEN_PATH    Set the bot token file path to TOKEN_PATH
        -s SERVER_PATH   Set the server id file path to SERVER_PATH

    Log Levels:
        [debug, info, warn, error, fatal]

    Default Options: [SpikeBot -l "%s" -t "%s" -s "%s"]
`
)

type programArgs struct {
	botToken string
	guildID  string
}

func main() {
	// Set the logrus logging formatter
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: logTimeStampFmt,
	})

	args := parseFlags()

	spike := startSpikeSession(args.botToken, args.guildID)
	defer spike.Close()

	spike.registerCommmands(cmds.CommandList)
	defer spike.deregisterAllCommands()

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
	var guildIDPath string
	flag.BoolVar(&printHelp, "h", false, "")
	flag.StringVar(&botTokenPath, "t", defaultBotTokenPath, "h")
	flag.StringVar(&guildIDPath, "g", defaultGuildIDPath, "")
	flag.StringVar(&logLevelStr, "l", defaultLogLevelStr, "")
	flag.Usage = func() {
		log.Fatalf(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath, defaultGuildIDPath)
	}
	flag.Parse()

	if printHelp {
		if printHelp {
			log.Infof(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath, defaultGuildIDPath)
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

	// Open and read guild ID from provided file path or default path
	guildIDFile, err := os.Open(guildIDPath)
	if err != nil {
		log.Fatalf("Failed to open server ID file path '%s'", guildIDPath)
	}
	guildID, err := io.ReadAll(guildIDFile)
	if err != nil {
		log.Fatalf("Failed to read from server ID file path '%s'", guildIDPath)
	}

	return &programArgs{
		botToken: string(botToken),
		guildID:  string(guildID),
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
	guildID            string
	registeredCommands map[string]*dg.ApplicationCommand
}

// Connects to the discord server and starts spike using the specified botToken
func startSpikeSession(botToken string, guildID string) *spikeSession {
	// Create a new Discord session using the provided bot token.
	session, err := dg.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Set the bot permission requirements
	session.Identify.Intents = dg.IntentGuildMessages | dg.IntentGuilds

	session.AddHandler(cmds.OnInteractionCreate)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	return &spikeSession{
		Session:            session,
		guildID:            guildID,
		registeredCommands: make(map[string]*dg.ApplicationCommand),
	}

}

func (s *spikeSession) registerCommmands(cmds []*dg.ApplicationCommand) {
	for _, cmd := range cmds {
		if _, ok := s.registeredCommands[cmd.ID]; ok {
			continue
		}
		regCmd, _ := s.ApplicationCommandCreate(s.State.User.ID, s.guildID, cmd)
		s.registeredCommands[regCmd.ID] = regCmd
	}
}

func (s *spikeSession) deregisterAllCommands() {
	for cmdID := range s.registeredCommands {
		s.ApplicationCommandDelete(s.State.User.ID, s.guildID, cmdID)
	}
	s.registeredCommands = make(map[string]*dg.ApplicationCommand)
}

// blocks until a kill signal is sent to the program such as "CTRL-C"
func waitForKillSig() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
