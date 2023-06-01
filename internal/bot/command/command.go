package command

import "github.com/bwmarrin/discordgo"

type Registerer interface {
	RegisterCommand(command discordgo.ApplicationCommand, fn HandleFunc)
	RegisterHandler(name string, fn HandleFunc)
	Commands() []*discordgo.ApplicationCommand
	Handlers() map[string]HandleFunc
}

type Command interface {
	Register(reg Registerer)
}

type HandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate) error

type commandRegisterer struct {
	cmds  []*discordgo.ApplicationCommand
	funcs map[string]HandleFunc
}

func NewRegisterer() Registerer {
	return &commandRegisterer{
		cmds:  []*discordgo.ApplicationCommand{},
		funcs: make(map[string]HandleFunc),
	}
}

func (r *commandRegisterer) RegisterCommand(command discordgo.ApplicationCommand, fn HandleFunc) {
	r.cmds = append(r.cmds, &command)
	r.funcs[command.Name] = fn
}

func (r *commandRegisterer) RegisterHandler(name string, fn HandleFunc) {
	r.funcs[name] = fn
}

func (r *commandRegisterer) Commands() []*discordgo.ApplicationCommand {
	return r.cmds
}

func (r *commandRegisterer) Handlers() map[string]HandleFunc {
	return r.funcs
}
