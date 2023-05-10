package study

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	sendFeedbackCmd = discordgo.ApplicationCommand{
		Name:        "피드백",
		Description: "발표자에게 피드백을 보냅니다.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "발표자",
				Description: "피드백을 받을 발표자를 선택해주세요.",
				Type:        discordgo.ApplicationCommandOptionUser,
				Required:    true,
			},
		},
	}
	feedbackTextInput = discordgo.TextInput{
		CustomID:    "feedback",
		Label:       "피드백",
		Style:       discordgo.TextInputParagraph,
		Placeholder: "피드백을 입력해주세요.",
		Required:    true,
		MaxLength:   1000,
		MinLength:   10,
	}
)

const FeedbackModalCustomID = "feedback-modal"

func (b *StudyBot) addSendFeedbackCmd() {
	b.hdr.AddCommand(sendFeedbackCmd, b.sendFeedbackCmdHandler)
	b.chdr.AddHandleFunc(FeedbackModalCustomID, b.feedbackSubmitHandler)
}

func (b *StudyBot) sendFeedbackCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		var presentor *discordgo.User

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "발표자":
				presentor = option.UserValue(s)
			}
		}

		if presentor == nil {
			return errors.Join(ErrRequiredArgs, errors.New("리뷰 대상자는 필수 입력 사항입니다."))
		}

		if presentor.Bot {
			return errors.New("봇은 리뷰 대상자로 지정할 수 없습니다.")
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: FeedbackModalCustomID,
				Title:    "피드백 작성",
				Flags:    discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "presentor-id",
								Label:       "발표자",
								Style:       discordgo.TextInputShort,
								Placeholder: "발표자의 ID 입니다. 임의로 변경하지 마세요.",
								Value:       presentor.ID,
								Required:    true,
								MaxLength:   20,
								MinLength:   1,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{feedbackTextInput},
					},
				},
			},
		})
	}

	err := cmd(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) feedbackSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var reviewer *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			reviewer = i.Member.User
		}

		if reviewer == nil {
			return errors.Join(ErrUserNotFound, errors.New("리뷰어 정보를 찾을 수 없습니다."))
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
				case "presentor-id":
					revieweeID = input.Value
				case "feedback":
					feedback = input.Value
				}
			}
		}

		if revieweeID == "" || feedback == "" {
			return errors.Join(ErrRequiredArgs, errors.New("리뷰 대상자의 아이디 또는 피드백 정보를 찾을 수 없습니다."))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := b.svc.SetReviewer(ctx, i.GuildID, reviewer.ID, revieweeID)
		if err != nil {
			return err
		}

		channel, err := s.UserChannelCreate(revieweeID)
		if err != nil {
			return err
		}

		_, err = s.ChannelMessageSend(channel.ID, feedback)
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
		b.sugar.Errorw(err.Error(), "event", "feedback-submit")
		_ = errorInteractionRespond(s, i, err)
	}
}
