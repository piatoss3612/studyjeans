package study

import (
	"time"

	"github.com/bwmarrin/discordgo"
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
