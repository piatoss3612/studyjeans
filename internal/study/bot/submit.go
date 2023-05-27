package bot

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
)

var submitContentCmd = discordgo.ApplicationCommand{
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

func (b *StudyBot) addSubmitContentCmd() {
	b.cmd.AddCommand(submitContentCmd, b.submitContentCmdHandler)
}

func (b *StudyBot) submitContentCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		_, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
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
		b.sugar.Errorw(err.Error(), "event", "submit-content")
		_ = errorInteractionRespond(s, i, err)
	}
}
