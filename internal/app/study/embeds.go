package study

import "github.com/bwmarrin/discordgo"

func EmbedTemplate(u *discordgo.User, title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       title,
		Description: description,
		Color:       16777215,
	}
}
