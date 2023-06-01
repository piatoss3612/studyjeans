package submit

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
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"go.uber.org/zap"
)

var cmd = discordgo.ApplicationCommand{
	Name:        "발표-자료-제출",
	Description: "발표 자료를 제출합니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "링크",
			Description: "발표 자료 링크를 입력해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

type submitCommand struct {
	svc service.Service
	pub msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewSubmitCommand(svc service.Service, pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &submitCommand{
		svc:   svc,
		pub:   pub,
		sugar: sugar,
	}
}

func (sc *submitCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, sc.submitContentCmdHandler)
}

func (sc *submitCommand) submitContentCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return study.ErrUserNotFound
		}

		var content string

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "링크":
				content = option.StringValue()
			}
		}

		if content == "" {
			return errors.Join(study.ErrRequiredArgs, errors.New("발표 자료 링크는 필수 입력 사항입니다"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _, err := sc.svc.UpdateRound(ctx, &service.UpdateParams{
			GuildID:    i.GuildID,
			MemberID:   user.ID,
			ContentURL: content,
		},
			service.SubmitMemberContent, service.ValidateToSubmitMemberContent)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: user.Mention(),
				Flags:   discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					EmbedTemplate(s.State.User, "제출 완료", "발표 자료가 제출되었습니다."),
				},
			},
		})
	}

	err := fn(s, i)
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go sc.publishEvent(evt)
		}()
		sc.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (sc *submitCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := sc.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			sc.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
