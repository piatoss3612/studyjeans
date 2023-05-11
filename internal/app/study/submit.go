package study

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
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
	b.hdr.AddCommand(submitContentCmd, b.submitContentCmdHandler)
}

func (b *StudyBot) submitContentCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		var content string

		for _, option := range i.ApplicationCommandData().Options {
			switch option.Name {
			case "링크":
				content = option.StringValue()
			}
		}

		if content == "" {
			return errors.Join(ErrRequiredArgs, errors.New("발표 자료 링크는 필수 입력 사항입니다"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := b.svc.SetMemberContent(ctx, i.GuildID, user.ID, content)
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

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "submit-content")
		_ = errorInteractionRespond(s, i, err)
	}
}
