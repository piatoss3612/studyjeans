package component

import "github.com/bwmarrin/discordgo"

type Registerer interface {
	Register(name string, fn HandleFunc)
	Handlers() map[string]HandleFunc
}

type HandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type componentRegisterer struct {
	funcs map[string]HandleFunc
}

func NewRegisterer() Registerer {
	return &componentRegisterer{
		funcs: make(map[string]HandleFunc),
	}
}

func (r *componentRegisterer) Register(name string, fn HandleFunc) {
	r.funcs[name] = fn
}

func (r *componentRegisterer) Handlers() map[string]HandleFunc {
	return r.funcs
}
