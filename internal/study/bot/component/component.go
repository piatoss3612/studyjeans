package component

import "github.com/bwmarrin/discordgo"

type Registerer interface {
	Register(component Component)
	Handlers() map[string]HandleFunc
}

type Component interface {
	N() string
	F() HandleFunc
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

func (r *componentRegisterer) Register(component Component) {
	r.funcs[component.N()] = component.F()
}

func (r *componentRegisterer) Handlers() map[string]HandleFunc {
	return r.funcs
}
