package study

import (
	"github.com/bwmarrin/discordgo"
)

type Handler interface {
	AddCommand(cmd discordgo.ApplicationCommand, handleFunc HandleFunc)
	GetHandleFunc(name string) (HandleFunc, bool)
	RegisterApplicationCommands(s *discordgo.Session) error
	RemoveApplicationCommands(s *discordgo.Session) error
}

type HandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type handlerImpl struct {
	handleFuncs    map[string]HandleFunc
	cmds           []*discordgo.ApplicationCommand
	registeredCmds []*discordgo.ApplicationCommand
}

func NewHandler() Handler {
	return &handlerImpl{
		handleFuncs:    make(map[string]HandleFunc),
		cmds:           []*discordgo.ApplicationCommand{},
		registeredCmds: []*discordgo.ApplicationCommand{},
	}
}

func (h *handlerImpl) AddCommand(cmd discordgo.ApplicationCommand, handleFunc HandleFunc) {
	h.handleFuncs[cmd.Name] = handleFunc
	h.cmds = append(h.cmds, &cmd)
}

func (h *handlerImpl) GetHandleFunc(name string) (HandleFunc, bool) {
	f, ok := h.handleFuncs[name]
	return f, ok
}

func (h *handlerImpl) RegisterApplicationCommands(s *discordgo.Session) error {
	for _, cmd := range h.cmds {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return err
		}

		h.registeredCmds = append(h.registeredCmds, cmd)
	}

	return nil
}

func (h *handlerImpl) RemoveApplicationCommands(s *discordgo.Session) error {
	for _, cmd := range h.registeredCmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
