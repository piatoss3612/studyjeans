package registration

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	registerCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록",
		Description: "발표자 정보를 등록합니다.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "이름",
				Description: "발표자의 이름을 입력해주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "주제",
				Description: "발표 주제를 입력해주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	unregisterCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록-취소",
		Description: "발표자 등록을 취소합니다.",
	}
)

func EmbedTemplate(u *discordgo.User, title, description string, url ...string) *discordgo.MessageEmbed {
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
