package study

import (
	"github.com/bwmarrin/discordgo"
)

type Handler interface {
	AddCommand(cmd discordgo.ApplicationCommand, handleFunc HandleFunc)
	GetHandleFunc(name string) (HandleFunc, bool)
	RegisterApplicationCommands() error
	RemoveApplicationCommands() error
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

// AddCommand implements Handler.
func (*handlerImpl) AddCommand(cmd discordgo.ApplicationCommand, handleFunc HandleFunc) {
	panic("unimplemented")
}

// GetHandleFunc implements Handler.
func (*handlerImpl) GetHandleFunc(name string) (HandleFunc, bool) {
	panic("unimplemented")
}

// RegisterApplicationCommands implements Handler.
func (*handlerImpl) RegisterApplicationCommands() error {
	panic("unimplemented")
}

// RemoveApplicationCommands implements Handler.
func (*handlerImpl) RemoveApplicationCommands() error {
	panic("unimplemented")
}
