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
		cmdSkill(s, i)
	case "guest":
		cmdGuest(s, i)
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
						Description: "User whose skill rank will be increased",
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
			{
				Name:        "guest",
				Description: "f",
				Type:        dg.ApplicationCommandOptionSubCommandGroup,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "set",
						Description: "Set the skill rank of a guest.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "The name of the guest whose skill will be set.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
							{
								Name:        "skill",
								Description: "The skill rank to which the guest will be set.",
								Type:        dg.ApplicationCommandOptionInteger,
								Required:    true,
								MinValue:    ptr(float64(0)),
								MaxValue:    float64(100),
							},
						},
					},
					{
						Name:        "increase",
						Description: "Increase the skill rank of a guest.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "The name of the guest whose skill will be increased.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
							{
								Name:        "amount",
								Description: "The amount by which to increase the skill rank.",
								Type:        dg.ApplicationCommandOptionInteger,
								Required:    true,
								MinValue:    ptr(float64(0)),
								MaxValue:    float64(100),
							},
						},
					},
					{
						Name:        "decrease",
						Description: "Decrease the skill rank of a guest.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "The name of the guest whose skill will be decreased.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
							{
								Name:        "amount",
								Description: "The amount by which to decrease the skill rank.",
								Type:        dg.ApplicationCommandOptionInteger,
								Required:    true,
								MinValue:    ptr(float64(0)),
								MaxValue:    float64(100),
							},
						},
					},
					{
						Name:        "show",
						Description: "Display the current skill rank of a guest.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "Guest whose rank will be shown.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
						},
					},
				},
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
				Name:        "show_all",
				Description: "Show the list of active players and their skill ranks",
				Type:        dg.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "guest",
				Description: "f",
				Type:        dg.ApplicationCommandOptionSubCommandGroup,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "add",
						Description: "Add a guest to the playing group.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "The name of the guest to add to the playing group.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
						},
					},
					{
						Name:        "remove",
						Description: "Remove a guest from the playing group.",
						Type:        dg.ApplicationCommandOptionSubCommand,
						Options: []*dg.ApplicationCommandOption{
							{
								Name:        "player",
								Description: "The name of the guest to remove from the playing group.",
								Type:        dg.ApplicationCommandOptionRole,
								Required:    true,
							},
						},
					},
				},
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
		Description: "Update Spike database with player's names and remove players that have left the server",
	},
	{
		Name:        "guest",
		Description: "Guest variations of other commands",
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add a new guest with a name and skill rank.",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "The name of the guest to add",
						Type:        dg.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "skill",
						Description: "The skill rank of the guest to add",
						Type:        dg.ApplicationCommandOptionInteger,
						Required:    true,
						MinValue:    ptr(float64(0)),
						MaxValue:    float64(100),
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove guest.",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "The name of the guest to remove.",
						Type:        dg.ApplicationCommandOptionRole,
						Required:    true,
					},
				},
			},
			{
				Name:        "rename",
				Description: "Change the name of a guest.",
				Type:        dg.ApplicationCommandOptionSubCommand,
				Options: []*dg.ApplicationCommandOption{
					{
						Name:        "player",
						Description: "Guest whose name will be changed.",
						Type:        dg.ApplicationCommandOptionRole,
						Required:    true,
					},
					{
						Name:        "name",
						Description: "The new name to be assigned to the guest.",
						Type:        dg.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
			{
				Name:        "show_all",
				Description: "Display all guests and their skill ranks.",
				Type:        dg.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

const helpMessage = "Spike Command Options:\n" +
	"```" + `
help
skill
	set
	increase
	decrease
	show
	show_all
	guest
		set
		increase
		decrease
		show
playing
	add
	remove
	clear
	show_all
	guest
		add
		remove
guest
	add
	remove
	rename
	show_all
teams
update_names
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
