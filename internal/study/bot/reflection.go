package bot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
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
	b.cmd.AddCommand(reflectionCmd, b.reflectionCmdHandler)
}

func (b *StudyBot) reflectionCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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

		// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// defer cancel()

		// // set sent reflection
		// reflectionChID, err := b.svc.SetSentReflection(ctx, i.GuildID, user.ID)
		// if err != nil {
		// 	return err
		// }

		// embed := &discordgo.MessageEmbed{
		// 	Author: &discordgo.MessageEmbedAuthor{
		// 		Name:    user.Username,
		// 		IconURL: user.AvatarURL(""),
		// 	},
		// 	Title: "발표회고",
		// 	Fields: []*discordgo.MessageEmbedField{
		// 		{
		// 			Name:  "내용",
		// 			Value: content,
		// 		},
		// 	},
		// 	Color:     0x00ffff,
		// 	Timestamp: time.Now().Format(time.RFC3339),
		// }

		// // send reflection
		// _, err = s.ChannelMessageSendEmbed(reflectionChID, embed)
		// if err != nil {
		// 	return err
		// }

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
