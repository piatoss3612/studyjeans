package ping

import (
	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/studyjeans/pkg/command"
)

type PingCommand struct {
}

func New() *PingCommand {
	return &PingCommand{}
}

func (c *PingCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "ping pong",
	}
}

func (c *PingCommand) HandleFunc() command.CommandHandleFunc {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "pong",
			},
		})
	}
}

// InteractionHandleFuncs implements command.Commander.
func (c *PingCommand) InteractionHandleFuncs() map[string]command.CommandHandleFunc {
	return nil
}

var _ command.Commander = (*PingCommand)(nil)
