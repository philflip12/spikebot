package commands

// This file contains the command definitions available for SpikeBot. These definitions are
// registered on the discord server at startup.

import (
	"fmt"

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

var (
	memberOption = &dg.ApplicationCommandOption{
		Name:        "name",
		Description: "Name of the member in the server",
		Type:        dg.ApplicationCommandOptionUser,
		Required:    true,
	}
	memberOptionNumber = func(number int) *dg.ApplicationCommandOption {
		return &dg.ApplicationCommandOption{
			Name:        fmt.Sprintf("name_%d", number),
			Description: "Name of a member in the server",
			Type:        dg.ApplicationCommandOptionUser,
			Required:    number == 1,
		}
	}
	multiMemberSelectOptions = func(count int) []*dg.ApplicationCommandOption {
		options := make([]*dg.ApplicationCommandOption, count)
		for i := 1; i <= count; i++ {
			options[i-1] = memberOptionNumber(i)
		}
		return options
	}
	guestOption = &dg.ApplicationCommandOption{
		Name:        "name",
		Description: "Name of the guest",
		Type:        dg.ApplicationCommandOptionRole,
		Required:    true,
	}
	guestOptionNumber = func(number int) *dg.ApplicationCommandOption {
		return &dg.ApplicationCommandOption{
			Name:        fmt.Sprintf("name_%d", number),
			Description: "Name of a guest",
			Type:        dg.ApplicationCommandOptionRole,
			Required:    number == 1,
		}
	}
	multiGuestSelectOptions = func(count int) []*dg.ApplicationCommandOption {
		options := make([]*dg.ApplicationCommandOption, count)
		for i := 1; i <= count; i++ {
			options[i-1] = guestOptionNumber(i)
		}
		return options
	}
	skillOption = &dg.ApplicationCommandOption{
		Name:        "skill",
		Description: "Skill rank to set",
		Type:        dg.ApplicationCommandOptionInteger,
		Required:    true,
		MinValue:    ptr(float64(0)),
		MaxValue:    99,
	}
	increaseOption = &dg.ApplicationCommandOption{
		Name:        "amount",
		Description: "Amount by which to increase",
		Type:        dg.ApplicationCommandOptionInteger,
		Required:    true,
		MinValue:    ptr(float64(1)),
		MaxValue:    99,
	}
	decreaseOption = &dg.ApplicationCommandOption{
		Name:        "amount",
		Description: "Amount by which to decrease",
		Type:        dg.ApplicationCommandOptionInteger,
		Required:    true,
		MinValue:    ptr(float64(1)),
		MaxValue:    99,
	}
)

// func OnInteractionEvent(s *dg.Session, m *dg.InteractionCreate, ...)

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
		Options: []*dg.ApplicationCommandOption{
			memberOption,
			skillOption,
		},
	}, {
		Name:        "increase",
		Description: "Increase the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{
			memberOption,
			increaseOption,
		},
	}, {
		Name:        "decrease",
		Description: "Decrease the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{
			memberOption,
			decreaseOption,
		},
	}, {
		Name:        "show",
		Description: "Display the skill rank of a player",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{
			memberOption,
		},
	}, {
		Name:        "show_all",
		Description: "Display the skill rank of all players",
		Type:        dg.ApplicationCommandOptionSubCommand,
	}, {
		Name:        "guest",
		Description: "Guest variations of skill commands",
		Type:        dg.ApplicationCommandOptionSubCommandGroup,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "set",
			Description: "Set the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{
				guestOption,
				skillOption,
			},
		}, {
			Name:        "increase",
			Description: "Increase the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{
				guestOption,
				increaseOption,
			},
		}, {
			Name:        "decrease",
			Description: "Decrease the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{
				guestOption,
				decreaseOption,
			},
		}, {
			Name:        "show",
			Description: "Display the skill rank of a guest",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options: []*dg.ApplicationCommandOption{
				guestOption},
		}},
	}},
}, {
	Name:        "playing",
	Description: "Commads relating to the list of active players",
	Options: []*dg.ApplicationCommandOption{{
		Name:        "add",
		Description: "Add players to the list of active players",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options:     multiMemberSelectOptions(24),
	}, {
		Name:        "remove",
		Description: "Remove players from the list of active players",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options:     multiMemberSelectOptions(24),
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
			Description: "Add guests to the playing group",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options:     multiGuestSelectOptions(24),
		}, {
			Name:        "remove",
			Description: "Remove guests from the playing group",
			Type:        dg.ApplicationCommandOptionSubCommand,
			Options:     multiGuestSelectOptions(24),
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
		Description: "Create a new guest",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{
			{
				Name:        "name",
				Description: "The name of the guest",
				Type:        dg.ApplicationCommandOptionString,
				Required:    true,
			},
			skillOption,
		},
	}, {
		Name:        "delete",
		Description: "Delete guest",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{
			guestOption,
		},
	}, {
		Name:        "rename",
		Description: "Change the name of a guest",
		Type:        dg.ApplicationCommandOptionSubCommand,
		Options: []*dg.ApplicationCommandOption{{
			Name:        "old_name",
			Description: "Guest whose name will be changed",
			Type:        dg.ApplicationCommandOptionRole,
			Required:    true,
		}, {
			Name:        "new_name",
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
