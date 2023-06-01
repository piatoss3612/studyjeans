package utils

import "github.com/bwmarrin/discordgo"

func GetGuildUserFromInteraction(i *discordgo.InteractionCreate) (user *discordgo.User) {
	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}
	return
}
