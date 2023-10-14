package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// CommandManager is an interface for discord slash command creation and deletion
type CommandManager interface {
	CommandCreate(guildID string, cmd *discordgo.ApplicationCommand) error
	CommandDelete(guildID string, cmdID string) error
}

// Commander is an interface for a discord slash command
type Commander interface {
	Command() *discordgo.ApplicationCommand
	HandleFunc() CommandHandleFunc
	InteractionHandleFuncs() map[string]CommandHandleFunc
}

// CommandHandleFunc is a function that handles a discord slash command or one of its interactions
type CommandHandleFunc func(*discordgo.Session, *discordgo.InteractionCreate) error

// CommandRegistry is a registry for discord slash commands
type CommandRegistry struct {
	m        CommandManager
	cmds     []*discordgo.ApplicationCommand
	handlers map[string]CommandHandleFunc
}

// NewCommandRegistry creates a new CommandRegistry
func NewCommandRegistry(m CommandManager) *CommandRegistry {
	return &CommandRegistry{
		m:        m,
		cmds:     make([]*discordgo.ApplicationCommand, 0),
		handlers: make(map[string]CommandHandleFunc),
	}
}

// RegisterCommand registers a discord slash command to the registry
func (r *CommandRegistry) RegisterCommand(c Commander) {
	cmd := c.Command()
	r.cmds = append(r.cmds, cmd)
	r.handlers[cmd.Name] = c.HandleFunc()
	for k, v := range c.InteractionHandleFuncs() {
		r.handlers[k] = v
	}
}

// RegisterCommands registers multiple discord slash commands to the registry
func (r *CommandRegistry) RegisterCommands(cs ...Commander) {
	for _, c := range cs {
		r.RegisterCommand(c)
	}
}

// CreateCommands creates the discord slash commands in the registry on discord
func (r *CommandRegistry) CreateCommands() error {
	for _, c := range r.cmds {
		err := r.m.CommandCreate("", c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Handle handles a discord slash command or one of its interactions
func (r *CommandRegistry) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var name string

	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		name = i.ApplicationCommandData().Name
	case discordgo.InteractionMessageComponent:
		name = i.MessageComponentData().CustomID
	case discordgo.InteractionModalSubmit:
		name = i.ModalSubmitData().CustomID
	default:
		return fmt.Errorf("interaction type %v not found", i.Type)
	}

	if h, ok := r.handlers[name]; ok {
		return h(s, i)
	}

	return fmt.Errorf("handler for interaction %s not found", name)
}

// DeleteCommands deletes the discord slash commands in the registry on discord
func (r *CommandRegistry) DeleteCommands() error {
	for _, c := range r.cmds {
		err := r.m.CommandDelete("", c.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
