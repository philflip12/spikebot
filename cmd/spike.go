package main

import (
	"flag"
	"io"
	"os"
	"os/signal"
	"syscall"

	dg "github.com/bwmarrin/discordgo"
	"github.com/philflip12/spikebot/internal/commands"
	log "github.com/sirupsen/logrus"
)

const (
	logTimeStampFmt     = "06-01-02 3:04:05 pm" // YY-MM-DD HH:MM:SS pm
	defaultBotTokenPath = "./.env/BotToken"
	defaultLogLevelStr  = "info"
	usageDialogFmtStr   = `
    SpikeBot [Options...]

    Options:
        -h               Print this help dialog
        -l LOG_LEVEL     Set program to print logs of LOG_LEVEL and higher
        -t TOKEN_PATH    Set the bot token file path to TOKEN_PATH

    Log Levels:
        [debug, info, warn, error, fatal]

    Default Options: [SpikeBot -l "%s" -t "%s"]
`
)

type programArgs struct {
	botToken string
}

func main() {
	// Set the logrus logging formatter
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: logTimeStampFmt,
	})

	args := parseFlags()

	session := startBotSession(args.botToken)
	defer session.Close()

	log.Info("Bot is now running. Press CTRL-C to exit")

	waitForKillSig()
}

// Reads the command line flags for running the "Spike" discord bot.
// Kills the program on error or help request.
func parseFlags() *programArgs {
	var printHelp bool
	var logLevelStr string
	var botTokenPath string
	flag.BoolVar(&printHelp, "h", false, "Print instructions to run this program and exit")
	flag.StringVar(&botTokenPath, "t", defaultBotTokenPath, "Bot Token")
	flag.StringVar(&logLevelStr, "l", defaultLogLevelStr, "Log Level [debug, info, warn, error, fatal]")
	flag.Usage = func() {
		log.Fatalf(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath)
	}
	flag.Parse()

	if printHelp {
		if printHelp {
			log.Infof(usageDialogFmtStr, defaultLogLevelStr, defaultBotTokenPath)
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

	return &programArgs{
		botToken: string(botToken),
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

// Connects to the discord server and starts spike using the specified botToken
func startBotSession(botToken string) *dg.Session {
	// Create a new Discord session using the provided bot token.
	session, err := dg.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Register the onMessageCreate func as a callback for onMessageCreate events.
	session.AddHandler(commands.OnMessageCreate)

	// Set the bot permission requirements
	session.Identify.Intents = dg.IntentGuildMessages | dg.IntentGuilds

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	return session
}

// blocks until a kill signal is sent to the program such as "CTRL-C"
func waitForKillSig() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
