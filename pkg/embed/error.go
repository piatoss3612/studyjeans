package embed

import "github.com/bwmarrin/discordgo"

func ErrorEmbed(err error) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: err.Error(),
		Color:       0xff0000,
	}
}
