package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var ErrInteractionNotFound = fmt.Errorf("interaction not found")

// Commander is an interface for a discord slash command
type Commander interface {
	Command() *discordgo.ApplicationCommand
	HandleFunc() CommandHandleFunc
	InteractionHandleFuncs() map[string]CommandHandleFunc
}

// ApplicationCommandManager is an interface for discord slash command creation and deletion
type ApplicationCommandManager interface {
	ApplicationCommandCreate(guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error)
	ApplicationCommandDelete(guildID string, cmdID string, options ...discordgo.RequestOption) error
}

// CommandHandleFunc is a function that handles a discord slash command or one of its interactions
type CommandHandleFunc func(*discordgo.Session, *discordgo.InteractionCreate) error

// CommandRegistry is a registry for discord slash commands
type CommandRegistry struct {
	m        ApplicationCommandManager
	cmds     []*discordgo.ApplicationCommand
	handlers map[string]CommandHandleFunc
}

// NewCommandRegistry creates a new CommandRegistry
func NewCommandRegistry(m ApplicationCommandManager) *CommandRegistry {
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
		_, err := r.m.ApplicationCommandCreate("", c)
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
	}

	if h, ok := r.handlers[name]; ok {
		return h(s, i)
	}

	return fmt.Errorf("%s: %s", ErrInteractionNotFound.Error(), name)
}

// DeleteCommands deletes the discord slash commands in the registry on discord
func (r *CommandRegistry) DeleteCommands() error {
	for _, c := range r.cmds {
		err := r.m.ApplicationCommandDelete("", c.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
