package study

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/utils"
)

var profileCmd = discordgo.ApplicationCommand{
	Name:        "프로필",
	Description: "발표 진스의 프로필을 보여줍니다.",
}

func (b *StudyBot) addProfileCmd() {
	b.hdr.AddCommand(profileCmd, b.profileCmdHandler)
}

func (b *StudyBot) profileCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	u := s.State.User
	createdAt, _ := utils.FormatSnowflakeToTime(u.ID)
	rebootedAt := utils.FormatRebootDate(b.startedAt)
	uptime := utils.FormatUptime(b.startedAt)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: u.Mention(),
			Embeds: []*discordgo.MessageEmbed{
				ProfileEmbed(u, "발표 진스의 프로필", createdAt, rebootedAt, uptime),
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "profile")
		_ = errorInteractionRespond(s, i, err)
	}
}

func ProfileEmbed(u *discordgo.User, title, createdAt, rebootedAt, uptime string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "이름",
				Value:  fmt.Sprintf("```%s```", u.Username),
				Inline: true,
			},
			{
				Name:   "생성일",
				Value:  fmt.Sprintf("```%s```", createdAt),
				Inline: true,
			},
			{
				Name:   "재부팅",
				Value:  fmt.Sprintf("```%s```", rebootedAt),
				Inline: true,
			},
			{
				Name:   "업타임",
				Value:  fmt.Sprintf("```%s```", uptime),
				Inline: true,
			},
			{
				Name:   "💻 개발자",
				Value:  fmt.Sprintf("```%s```", "piatoss3612"),
				Inline: true,
			},
			{
				Name:  "📝 소스코드",
				Value: fmt.Sprintf("```%s```", "https://github.com/piatoss3612/presentation-helper-bot"),
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: u.AvatarURL("256"),
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     16777215,
	}
}
