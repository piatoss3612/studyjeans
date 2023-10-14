package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
	cmds     []*discordgo.ApplicationCommand
	handlers map[string]CommandHandleFunc
}

// NewCommandRegistry creates a new CommandRegistry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
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

// Handler returns the handler of the command with the given name
func (r *CommandRegistry) Handler(name string) (CommandHandleFunc, bool) {
	h, ok := r.handlers[name]
	return h, ok
}

// Commands returns the commands of the registry
func (r *CommandRegistry) Commands() []*discordgo.ApplicationCommand {
	return r.cmds
}

// Handlers returns the handlers of the registry
func (r *CommandRegistry) Handlers() map[string]CommandHandleFunc {
	return r.handlers
}

// CommandManager is a manager for discord slash commands
type CommandManager struct {
	s *discordgo.Session
	r *CommandRegistry
}

// NewCommandManager creates a new CommandManager
func NewCommandManager(s *discordgo.Session, r *CommandRegistry) *CommandManager {
	return &CommandManager{
		s: s,
		r: r,
	}
}

// CommandRegistry returns the command registry of the manager
func (m *CommandManager) CommandRegistry() *CommandRegistry {
	return m.r
}

// SetCommandRegistry sets the command registry of the manager
func (m *CommandManager) SetCommandRegistry(r *CommandRegistry) {
	m.r = r
}

// CreateCommands creates the discord slash commands
func (m *CommandManager) CreateCommands(guildID string) error {
	cmds := m.r.Commands()
	for _, cmd := range cmds {
		_, err := m.s.ApplicationCommandCreate(m.s.State.User.ID, guildID, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// Handle handles a discord slash command or one of its interactions
func (m *CommandManager) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	if h, ok := m.r.Handler(name); ok {
		return h(s, i)
	}

	return fmt.Errorf("handler for interaction %s not found", name)
}

// DeleteCommands deletes the discord slash commands
func (m *CommandManager) DeleteCommands(guildID string) error {
	cmds := m.r.Commands()
	for _, cmd := range cmds {
		err := m.s.ApplicationCommandDelete(m.s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
