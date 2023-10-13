package command

import "github.com/bwmarrin/discordgo"

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
	s        *discordgo.Session
	cmds     []*discordgo.ApplicationCommand
	handlers map[string]CommandHandleFunc
}

// NewCommandRegistry creates a new CommandRegistry
func NewCommandRegistry(s *discordgo.Session) *CommandRegistry {
	return &CommandRegistry{
		s:        s,
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
		_, err := r.s.ApplicationCommandCreate(r.s.State.User.ID, "", c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Handle handles a discord slash command or one of its interactions
func (r *CommandRegistry) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if h, ok := r.handlers[i.ApplicationCommandData().Name]; ok {
		return h(s, i)
	}

	return nil
}

// DeleteCommands deletes the discord slash commands in the registry on discord
func (r *CommandRegistry) DeleteCommands() error {
	for _, c := range r.cmds {
		err := r.s.ApplicationCommandDelete(r.s.State.User.ID, "", c.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
