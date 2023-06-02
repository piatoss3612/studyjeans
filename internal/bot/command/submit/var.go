package submit

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var cmd = discordgo.ApplicationCommand{
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

func submitEmbed(u *discordgo.User, title, description string, url ...string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       title,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       16777215,
	}

	if len(url) > 0 {
		embed.URL = url[0]
	}

	return embed
}
