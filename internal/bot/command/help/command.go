package help

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/study"
)

type helpCommand struct {
}

func NewHelpCommand() command.Command {
	return &helpCommand{}
}

func (h *helpCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, h.help)
	reg.RegisterHandler(selectMenu.CustomID, h.selectHelpMenu)
}

// show help embed
func (h *helpCommand) help(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{HelpIntroEmbed(s.State.User)},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						selectMenu,
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{button},
				},
			},
		},
	}

	return s.InteractionRespond(i.Interaction, response)
}

func (h *helpCommand) selectHelpMenu(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var embed *discordgo.MessageEmbed

	data := i.MessageComponentData().Values
	if len(data) == 0 {
		return errors.Join(study.ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
	}

	switch data[0] {
	case "default":
		embed = HelpDefaultEmbed(s.State.User)
	case "study":
		embed = HelpStudyEmbed(s.State.User)
	default:
		return errors.Join(study.ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						selectMenu,
					},
				},
			},
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}

	return s.InteractionRespond(i.Interaction, response)
}
