package help

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/event/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"go.uber.org/zap"
)

type helpCommand struct {
	pub msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewHelpCommand(pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &helpCommand{
		pub:   pub,
		sugar: sugar,
	}
}

func (h *helpCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, h.helpCmdHandler)
	reg.RegisterHandler(selectMenu.CustomID, h.helpSelectMenuHandler)
}

func (h *helpCommand) helpCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
	})
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go h.publishEvent(evt)
		}()
		h.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (h *helpCommand) helpSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

	err := fn(s, i)
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go h.publishEvent(evt)
		}()
		h.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (cmd *helpCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := cmd.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			cmd.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
