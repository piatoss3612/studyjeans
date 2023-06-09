package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Handler interface {
	Handle(name string, s *discordgo.Session, i *discordgo.InteractionCreate) error
}

type handler struct {
	funcs map[string]HandleFunc
}

func NewHandler(funcs map[string]HandleFunc) Handler {
	return &handler{
		funcs: funcs,
	}
}

func (h *handler) Handle(name string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	fn, ok := h.funcs[name]
	if !ok {
		return fmt.Errorf("command %s not found", name)
	}
	return fn(s, i)
}
