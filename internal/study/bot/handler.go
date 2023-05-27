package bot

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler interface {
	AddCommand(cmd discordgo.ApplicationCommand, f HandleFunc)
	GetHandleFunc(name string) (HandleFunc, bool)
	RegisterApplicationCommands(s *discordgo.Session) error
	RemoveApplicationCommands(s *discordgo.Session) error
}

type HandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type commandHandler struct {
	handleFuncs    map[string]HandleFunc
	cmds           []*discordgo.ApplicationCommand
	registeredCmds []*discordgo.ApplicationCommand
}

func NewCommandHandler() CommandHandler {
	return &commandHandler{
		handleFuncs:    make(map[string]HandleFunc),
		cmds:           []*discordgo.ApplicationCommand{},
		registeredCmds: []*discordgo.ApplicationCommand{},
	}
}

func (h *commandHandler) AddCommand(cmd discordgo.ApplicationCommand, f HandleFunc) {
	h.handleFuncs[cmd.Name] = f
	h.cmds = append(h.cmds, &cmd)
}

func (h *commandHandler) GetHandleFunc(name string) (HandleFunc, bool) {
	f, ok := h.handleFuncs[name]
	return f, ok
}

func (h *commandHandler) RegisterApplicationCommands(s *discordgo.Session) error {
	for _, cmd := range h.cmds {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return err
		}

		h.registeredCmds = append(h.registeredCmds, cmd)
	}

	return nil
}

func (h *commandHandler) RemoveApplicationCommands(s *discordgo.Session) error {
	for _, cmd := range h.registeredCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

type ComponentHandler interface {
	AddComponent(name string, f HandleFunc)
	GetHandleFunc(name string) (HandleFunc, bool)
}

type componentCommandHandler struct {
	handleFuncs map[string]HandleFunc
}

func NewComponentHandler() ComponentHandler {
	return &componentCommandHandler{
		handleFuncs: make(map[string]HandleFunc),
	}
}

func (h *componentCommandHandler) AddComponent(name string, f HandleFunc) {
	h.handleFuncs[name] = f
}

func (h *componentCommandHandler) GetHandleFunc(name string) (HandleFunc, bool) {
	f, ok := h.handleFuncs[name]
	return f, ok
}
