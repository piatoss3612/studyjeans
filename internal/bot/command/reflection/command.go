package reflection

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

type reflectionCommand struct {
	svc service.Service
	pub msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewReflectionCommand(svc service.Service, pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &reflectionCommand{
		svc:   svc,
		pub:   pub,
		sugar: sugar,
	}
}

func (rc *reflectionCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, nil)
}

func (rc *reflectionCommand) reflectionCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		// user should be in guild
		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return study.ErrUserNotFound
		}

		content := i.ApplicationCommandData().Options[0].StringValue()

		// content should not be empty
		if content == "" {
			return errors.Join(study.ErrRequiredArgs, errors.New("회고 내용은 필수입니다"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// set sent reflection
		gs, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
			GuildID:  i.GuildID,
			MemberID: user.ID,
		},
			service.SetSentReflection, service.ValidateToSetSendReflection)
		if err != nil {
			return err
		}

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    user.Username,
				IconURL: user.AvatarURL(""),
			},
			Title: "발표회고",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "내용",
					Value: content,
				},
			},
			Color:     0x00ffff,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		// send reflection
		_, err = s.ChannelMessageSendEmbed(gs.ReflectionChannelID, embed)
		if err != nil {
			return err
		}

		// send success message
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "회고가 성공적으로 전송되었습니다.",
				Flags:   discordgo.MessageFlagsEphemeral,
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

			go rc.publishEvent(evt)
		}()
		rc.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (rc *reflectionCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := rc.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			rc.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
