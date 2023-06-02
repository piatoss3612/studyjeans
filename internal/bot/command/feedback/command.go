package feedback

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/study"
	"github.com/piatoss3612/my-study-bot/internal/study/service"
	"github.com/piatoss3612/my-study-bot/internal/utils"
	"go.uber.org/zap"
)

type feedbackCommand struct {
	svc service.Service

	sugar *zap.SugaredLogger
}

func NewFeedbackCommand(svc service.Service, sugar *zap.SugaredLogger) command.Command {
	return &feedbackCommand{
		svc:   svc,
		sugar: sugar,
	}
}

func (fc *feedbackCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(cmd, fc.showSendFeedbackModal)
	reg.RegisterHandler(feedbackModalCustomID, fc.sendFeedback)
}

// show send feedback modal
func (fc *feedbackCommand) showSendFeedbackModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// command should be used in guild
	reviewer := utils.GetGuildUserFromInteraction(i)
	if reviewer == nil {
		return study.ErrUserNotFound
	}

	// check speaker
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

	// reviewer can't feedback to yourself
	if speaker.ID == reviewer.ID {
		return study.ErrFeedbackYourself
	}

	// show modal
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

// send feedback
func (fc *feedbackCommand) sendFeedback(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// command should be used in guild
	reviewer := utils.GetGuildUserFromInteraction(i)
	if reviewer == nil {
		return errors.Join(study.ErrUserNotFound, errors.New("리뷰어 정보를 찾을 수 없습니다"))
	}

	data := i.ModalSubmitData()

	var speakerID, feedback string

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
				speakerID = input.Value
			case "feedback":
				feedback = input.Value
			}
		}
	}

	if speakerID == "" || feedback == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("리뷰 대상자의 아이디 또는 피드백 정보를 찾을 수 없습니다"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// set reviewer id
	_, _, err := fc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		ReviewerID: reviewer.ID,
		RevieweeID: speakerID,
	}, service.SetReviewer, service.ValidateToSetReviewer)
	if err != nil {
		return err
	}

	// create dm channel and send feedback
	channel, err := s.UserChannelCreate(speakerID)
	if err != nil {
		return err
	}

	embed := feedbackEmbed(s.State.User, feedback)

	_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "피드백이 전송되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
