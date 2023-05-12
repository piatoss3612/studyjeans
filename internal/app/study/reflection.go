package study

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

var reflectionCmd = discordgo.ApplicationCommand{
	Name:        "발표회고",
	Description: "발표회고를 작성합니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "내용",
			Description: "발표회고 내용을 입력해주세요.",
			Required:    true,
		},
	},
}

func (b *StudyBot) addReflectionCmd() {
	b.hdr.AddCommand(reflectionCmd, b.reflectionCmdHandler)
}

func (b *StudyBot) reflectionCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var user *discordgo.User

		// user should be in guild
		if i.Member != nil && i.Member.User != nil {
			user = i.Member.User
		}

		if user == nil {
			return ErrUserNotFound
		}

		content := i.ApplicationCommandData().Options[0].StringValue()

		// content should not be empty
		if content == "" {
			return errors.Join(ErrRequiredArgs, errors.New("회고 내용은 필수입니다"))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// set sent reflection
		reflectionChID, err := b.svc.SetSentReflection(ctx, i.GuildID, user.ID)
		if err != nil {
			return err
		}

		// TODO: create embed

		// send reflection
		_, err = s.ChannelMessageSend(reflectionChID, content)
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

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "reflection")
		_ = errorInteractionRespond(s, i, err)
	}
}
