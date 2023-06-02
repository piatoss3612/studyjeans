package reflection

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var cmd = discordgo.ApplicationCommand{
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

func reflectionEmbed(u *discordgo.User, content string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
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
}
