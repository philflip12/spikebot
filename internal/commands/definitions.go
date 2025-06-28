package commands

// This file contains the command definitions available for SpikeBot. These definitions are
// registered on the discord server at startup.

import (
	dg "github.com/bwmarrin/discordgo"
	rsp "github.com/philflip12/spikebot/internal/responder"
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
	case "continue":
		cmdContinue(s, i)
	case "playing":
		cmdPlay(s, i)
	case "skill":
		cmdSkill(s, i)
	case "guest":
		cmdGuest(s, i)
	case "update_names":
		cmdUpdateNames(s, i)
	case "teams":
		cmdTeams(s, i)
	case "redo":
		cmdRedoTeams(s, i)
	}
}

// func OnInteractionEvent(s *dg.Session, m *dg.InteractionCreate)

var CommandList = []*dg.ApplicationCommand{{
	Name:        "help",
	Description: "Print all spike commands",
}, {
	Name:        "continue",
	Description: "Continue printing output from the previous command",
}, {
	Name:        "skill",
	Description: "Commands relating to players' skill ranks",
	Options: []*dg.ApplicationCommandOption{{
		Name:        "set",
		Description: "Overwrite the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "User whose rank will be set",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}, {
			Name:        "skill",
			Description: "New skill rank to set",
			Type:        dg.ApplicationCommandOptionInteger,
			Required:    true,
			MinValue:    ptr(float64(0)),
			MaxValue:    100,
		}},
	}, {
		Name:        "increase",
		Description: "Increase the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "User whose skill rank will be increased",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}, {
			Name:        "amount",
			Description: "The amount by which to increase the skill rank",
			Type:        dg.ApplicationCommandOptionInteger,
			Required:    true,
			MinValue:    ptr(float64(1)),
			MaxValue:    100,
		}},
	}, {
		Name:        "decrease",
		Description: "Decrease the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "User whose rank will be decreased",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}, {
			Name:        "amount",
			Description: "The amount by which to decrease the skill rank",
			Type:        dg.ApplicationCommandOptionInteger,
			Required:    true,
			MinValue:    ptr(float64(1)),
			MaxValue:    100,
		}},
	}, {
		Name:        "show",
		Description: "Display the current skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "User whose rank will be shown",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}},
	}, {
		Name:        "show_all",
		Description: "Display the current skill rank of all players",
		Type:        dg.ApplicationCommandOptionSubCommand,
	}, {
		Name:        "guest",
		Description: "Guest variations of skill commands",
		Type:        dg.ApplicationCommandOptionSubCommandGroup,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "set",
			Description: "Set the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "player",
				Description: "The name of the guest whose skill will be set",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}, {
				Name:        "skill",
				Description: "The skill rank to which the guest will be set",
				Type:        dg.ApplicationCommandOptionInteger,
				Required:    true,
				MinValue:    ptr(float64(0)),
				MaxValue:    float64(100),
			}},
		}, {
			Name:        "increase",
			Description: "Increase the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "player",
				Description: "The name of the guest whose skill will be increased",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}, {
				Name:        "amount",
				Description: "The amount by which to increase the skill rank",
				Type:        dg.ApplicationCommandOptionInteger,
				Required:    true,
				MinValue:    ptr(float64(0)),
				MaxValue:    float64(100),
			}},
		}, {
			Name:        "decrease",
			Description: "Decrease the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "player",
				Description: "The name of the guest whose skill will be decreased",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}, {
				Name:        "amount",
				Description: "The amount by which to decrease the skill rank",
				Type:        dg.ApplicationCommandOptionInteger,
				Required:    true,
				MinValue:    ptr(float64(0)),
				MaxValue:    float64(100),
			}},
		}, {
			Name:        "show",
			Description: "Display the current skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "player",
				Description: "Guest whose rank will be shown",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}},
		}},
	}},
}, {
	Name:        "playing",
	Description: "Commads relating to the list of active players",
	Options: []*dg.ApplicationCommandOption{{
		Name:        "add",
		Description: "Add a player to the list of active players",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player_1",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}, {
			Name:        "player_2",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_3",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_4",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_5",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_6",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_7",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_8",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_9",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_10",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_11",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_12",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_13",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_14",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_15",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_16",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_17",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_18",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_19",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_20",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_21",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_22",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_23",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_24",
			Description: "User to add to the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}},
	}, {
		Name:        "remove",
		Description: "Remove a player from the list of active players",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player_1",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    true,
		}, {
			Name:        "player_2",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_3",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_4",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_5",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_6",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_7",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_8",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_9",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_10",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_11",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_12",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_13",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_14",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_15",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_16",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_17",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_18",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_19",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_20",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_21",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_22",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_23",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}, {
			Name:        "player_24",
			Description: "User to remove from the playing group",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    false,
		}},
	}, {
		Name:        "clear",
		Description: "Clear the list of active players",
		Type:        dg.ApplicationCommandOptionSubCommand,
	}, {
		Name:        "show_all",
		Description: "Show the list of active players and their skill ranks",
		Type:        dg.ApplicationCommandOptionSubCommand,
	}, {
		Name:        "guest",
		Description: "Guest variations of playing commands",
		Type:        dg.ApplicationCommandOptionSubCommandGroup,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "add",
			Description: "Add a guest to the playing group",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "guest_1",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}, {
				Name:        "guest_2",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_3",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_4",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_5",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_6",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_7",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_8",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_9",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_10",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_11",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_12",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_13",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_14",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_15",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_16",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_17",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_18",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_19",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_20",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_21",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_22",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_23",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_24",
				Description: "Guest to add to the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}},
		}, {
			Name:        "remove",
			Description: "Remove a guest from the playing group",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{{
				Name:        "guest_1",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    true,
			}, {
				Name:        "guest_2",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_3",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_4",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_5",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_6",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_7",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_8",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_9",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_10",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_11",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_12",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_13",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_14",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_15",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_16",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_17",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_18",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_19",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_20",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_21",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_22",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_23",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}, {
				Name:        "guest_24",
				Description: "Guest to remove from the playing group",
				Type:        dg.ApplicationCommandOptionRole,
				Required:    false,
			}},
		}},
	}},
}, {
	Name:        "teams",
	Description: "Create teams based on the list of players currently playing",
	Options: []*dg.ApplicationCommandOption{{
		Name:        "count",
		Description: "Number of teams to create",
		Type:        dg.ApplicationCommandOptionInteger,
		Required:    true,
		MinValue:    ptr(float64(2)),
	}, {
		Name:        "max_skill_gap",
		Description: "The largest allowable skill gap between the strongest and weakest created teams, defaults to 5",
		Type:        dg.ApplicationCommandOptionInteger,
		Required:    false,
	}},
}, {
	Name:        "redo",
	Description: "Create teams in the same way as the last call to /teams",
}, {
	Name:        "update_names",
	Description: "Update Spike database with player's names and remove players that have left the server",
}, {
	Name:        "guest",
	Description: "Commands for managing guests",
	Options: []*dg.ApplicationCommandOption{{
		Name:        "create",
		Description: "Create a new guest with a name and skill rank",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "The name of the guest to add",
			Type:        dg.ApplicationCommandOptionString,
			Required:    true,
		}, {
			Name:        "skill",
			Description: "The skill rank of the guest to add",
			Type:        dg.ApplicationCommandOptionInteger,
			Required:    true,
			MinValue:    ptr(float64(0)),
			MaxValue:    float64(100),
		}},
	}, {
		Name:        "delete",
		Description: "Delete guest",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "The name of the guest to remove",
			Type:        dg.ApplicationCommandOptionRole,
			Required:    true,
		}},
	}, {
		Name:        "rename",
		Description: "Change the name of a guest",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "player",
			Description: "Guest whose name will be changed",
			Type:        dg.ApplicationCommandOptionRole,
			Required:    true,
		}, {
			Name:        "name",
			Description: "The new name to be assigned to the guest",
			Type:        dg.ApplicationCommandOptionString,
			Required:    true,
		}},
	}, {
		Name:        "show_all",
		Description: "Display all guests and their skill ranks",
		Type:        dg.ApplicationCommandOptionSubCommand,
	}},
}}

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
	create
	delete
	rename
	show_all
teams
update_names
` + "```"

func cmdHelp(s *dg.Session, i *dg.InteractionCreate) {
	rsp.InteractionRespond(s, i, helpMessage)
}

func cmdContinue(s *dg.Session, i *dg.InteractionCreate) {
	if err := rsp.InteractionContinue(s, i); err == rsp.ErrNoResponseContinuation {
		rsp.InteractionRespond(s, i, err.Error())
	}
}

func ptr[T any](value T) *T {
	return &value
}
