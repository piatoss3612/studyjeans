package feedback

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

type feedbackCommand struct {
	svc service.Service
	pub msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewFeedbackCommand(svc service.Service, pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &feedbackCommand{
		svc:   svc,
		pub:   pub,
		sugar: sugar,
	}
}

func (fc *feedbackCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, fc.sendFeedbackCmdHandler)
	reg.RegisterHandler(feedbackModalCustomID, fc.feedbackSubmitHandler)
}

func (fc *feedbackCommand) sendFeedbackCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return study.ErrUserNotFound
		}

		var speaker *discordgo.User

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "발표자":
				speaker = option.UserValue(s)
			}
		}

		if speaker == nil {
			return errors.Join(study.ErrRequiredArgs, errors.New("리뷰 대상자는 필수 입력 사항입니다"))
		}

		if speaker.Bot {
			return errors.New("봇은 리뷰 대상자로 지정할 수 없습니다")
		}

		if speaker.ID == user.ID {
			return study.ErrFeedbackYourself
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: feedbackModalCustomID,
				Title:    "피드백 작성",
				Flags:    discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "speaker-id",
								Label:       "발표자",
								Style:       discordgo.TextInputShort,
								Placeholder: "발표자의 ID 입니다. 임의로 변경하지 마세요.",
								Value:       speaker.ID,
								Required:    true,
								MaxLength:   20,
								MinLength:   1,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{textInput},
					},
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		go func() {
			evt := &event.ErrorEvent{
				T: "study.error",
				D: fmt.Sprintf("%s: %s", i.ApplicationCommandData().Name, err.Error()),
				C: time.Now(),
			}

			go fc.publishEvent(evt)
		}()
		fc.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (fc *feedbackCommand) feedbackSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var reviewer *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			reviewer = i.Member.User
		}

		if reviewer == nil {
			return errors.Join(study.ErrUserNotFound, errors.New("리뷰어 정보를 찾을 수 없습니다"))
		}

		data := i.ModalSubmitData()

		var revieweeID, feedback string

		for _, c := range data.Components {
			row, ok := c.(*discordgo.ActionsRow)
			if !ok {
				continue
			}

			for _, rc := range row.Components {
				input, ok := rc.(*discordgo.TextInput)
				if !ok {
					continue
				}

				switch input.CustomID {
				case "speaker-id":
					revieweeID = input.Value
				case "feedback":
					feedback = input.Value
				}
			}
		}

		if revieweeID == "" || feedback == "" {
			return errors.Join(study.ErrRequiredArgs, errors.New("리뷰 대상자의 아이디 또는 피드백 정보를 찾을 수 없습니다"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _, err := fc.svc.UpdateRound(ctx, &service.UpdateParams{
			GuildID:    i.GuildID,
			ReviewerID: reviewer.ID,
			RevieweeID: revieweeID,
		}, service.SetReviewer, service.ValidateToSetReviewer)
		if err != nil {
			return err
		}

		channel, err := s.UserChannelCreate(revieweeID)
		if err != nil {
			return err
		}

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "익명",
				IconURL: s.State.User.AvatarURL(""),
			},
			Title:       "피드백",
			Description: feedback,
			Color:       0x00ff00,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "피드백이 전송되었습니다.",
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

			go fc.publishEvent(evt)
		}()
		fc.sugar.Errorw(err.Error(), "event", i.ApplicationCommandData().Name)
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
			},
		})
	}
}

func (fc *feedbackCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := fc.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			fc.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
