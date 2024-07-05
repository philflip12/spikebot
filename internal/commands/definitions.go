package commands

// This file contains the command definitions available for SpikeBot. These definitions are
// registered on the discord server at startup.

import (
	dg "github.com/bwmarrin/discordgo"
)

// onMessageCreate is called every time a message is created on any channel the bot has access to.
// It is added as a callback by 'AddHandler'
func OnInteractionCreate(s *dg.Session, i *dg.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "help":
		cmdHelp(s, i)
	case "play":
		cmdPlay(s, i)
	case "rank":
		cmdRank(s, i)
	case "last_active":
		cmdLastActive(s, i)
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
		Name:        "rank",
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
		Name:        "play",
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
				Name:        "del",
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
				MinValue:    ptr(float64(1)),
			},
		},
	},
}

const helpMessage = "" +
	"Spike Command Options:\n" +
	"\t/help\n" +
	"\t/play\n" +
	"\t\tadd\n" +
	"\t\tdel\n" +
	"\t\tclear\n" +
	"\t\tshow\n" +
	"\t/rank\n" +
	"\t\tset\n" +
	"\t\tadd\n" +
	"\t\tsub\n" +
	"\t\tshow\n" +
	"\t/last_active\n" +
	"\t/teams"

func cmdHelp(s *dg.Session, i *dg.InteractionCreate) {
	interactionRespond(s, i, helpMessage)
}

func ptr[T any](value T) *T {
	return &value
}

// teams
//

// last active @name
