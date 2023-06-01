package command

import "github.com/bwmarrin/discordgo"

type Registerer interface {
	Register(command Command)
	Commands() []*discordgo.ApplicationCommand
	Handlers() map[string]HandleFunc
}

type Command interface {
	C() discordgo.ApplicationCommand
	F() HandleFunc
}

type HandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

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

func (r *commandRegisterer) Register(command Command) {
	cmd := command.C()
	fn := command.F()

	r.cmds = append(r.cmds, &cmd)
	r.funcs[cmd.Name] = fn
}

func (r *commandRegisterer) Commands() []*discordgo.ApplicationCommand {
	return r.cmds
}

func (r *commandRegisterer) Handlers() map[string]HandleFunc {
	return r.funcs
}
