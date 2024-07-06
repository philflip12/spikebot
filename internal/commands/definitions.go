package commands

// This file contains the command definitions available for SpikeBot. These definitions are
// registered on the discord server at startup.

import (
	dg "github.com/bwmarrin/discordgo"
)

var channelMap map[string]struct{}

func init() {
	channelMap = map[string]struct{}{}
}

// should not be called after adding handles which use the ChannelIDs map.
func SetChannelIDs(channelIDs []string) {
	for _, channelID := range channelIDs {
		channelMap[channelID] = struct{}{}
	}
}

// OnInteractionCreate is called every time an interaction is created on any server the bot has
// registered commands to.
// It is added as a callback by 'discordgo.Session.AddHandler'
func OnInteractionCreate(s *dg.Session, i *dg.InteractionCreate) {
	if _, ok := channelMap[i.ChannelID]; !ok {
		// Ignore commands sent from non-added channels
		return
	}
	switch i.ApplicationCommandData().Name {
	case "help":
		cmdHelp(s, i)
	case "playing":
		cmdPlay(s, i)
	case "skill":
		cmdRank(s, i)
	case "last_active":
		cmdLastActive(s, i)
	case "update_names":
		cmdUpdateNames(s, i)
	case "teams":
		cmdTeams(s, i)
	}
}

// func OnInteractionEvent(s *dg.Session, m *dg.InteractionCreate)

var CommandList = []*dg.ApplicationCommand{
	{
		Name:        "help",
		Description: "Print all spike commands",
	},
	{
		Name:        "skill",
		Description: "Commands relating to players' skill ranks",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "set",
				Description: "Overwrite the skill rank of a player",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User whose rank will be set",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
					{
						Name:        "skill",
						Description: "New skill rank to set",
						Type:        dg.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    ptr(float64(0)),
						MaxValue:    100,
					},
				},
			},
			{
				Name:        "increase",
				Description: "Increase the skill rank of a player",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User whose rank will be increased",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
					{
						Name:        "amount",
						Description: "The amount by which to increase the skill rank",
						Type:        dg.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    ptr(float64(1)),
						MaxValue:    100,
					},
				},
			},
			{
				Name:        "decrease",
				Description: "Decrease the skill rank of a player",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User whose rank will be decreased",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
					{
						Name:        "amount",
						Description: "The amount by which to decrease the skill rank",
						Type:        dg.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    ptr(float64(1)),
						MaxValue:    100,
					},
				},
			},
			{
				Name:        "show",
				Description: "Display the current skill rank of a player",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User whose rank will be shown",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Name:        "show_all",
				Description: "Display the current skill rank of all players",
				Type:        dg.ApplicationCommandOptionSubCommand,
			},
		},
	},
	{
		Name:        "playing",
		Description: "Commads relating to the list of active players",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add a player to the list of active players",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User to add to the list",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a player from the list of active players",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "User to remove from the list",
						Type:        dg.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Name:        "clear",
				Description: "Clear the list of active players",
				Type:        dg.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "show",
				Description: "Show the list of active players and their skill ranks",
				Type:        dg.ApplicationCommandOptionSubCommand,
			},
		},
	},
	{
		Name:        "last_active",
		Description: "Print the last time a user was active",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "player",
				Description: "User whose last activity time information will be shown",
				Type:        dg.ApplicationCommandOptionUser,
				Required:    true,
			},
		},
	},
	{
		Name:        "teams",
		Description: "Create teams based on the list of players currently playing",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "count",
				Description: "Number of teams to create",
				Type:        dg.ApplicationCommandOptionInteger,
				Required:    true,
				MinValue:    ptr(float64(2)),
			},
			{
				Name:        "max_skill_gap",
				Description: "The largest allowable skill gap between the strongest and weakest created teams, defaults to 20",
				Type:        dg.ApplicationCommandOptionInteger,
				Required:    false,
			},
		},
	},
	{
		Name:        "update_names",
		Description: "Update Spike database with players names",
	},
}

const helpMessage = "Spike Command Options:\n" +
	"```" + `
help
playing
	show
	add
	remove
	clear
skill
	show
	set
	increase
	decrease
	show_all
teams
update_names
last_active
` + "```"

func cmdHelp(s *dg.Session, i *dg.InteractionCreate) {
	interactionRespond(s, i, helpMessage)
}

func ptr[T any](value T) *T {
	return &value
}

// teams
//

// last active @name
