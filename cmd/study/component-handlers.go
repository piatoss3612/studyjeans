package main

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

func (b *StudyBot) helpSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var embed *discordgo.MessageEmbed

	data := i.MessageComponentData().Values

	var err error

	if len(data) == 0 {
		err = errors.New("옵션을 찾을 수 없습니다.")
	} else {
		switch data[0] {
		case "default":
			embed = HelpDefaultEmbed(s.State.User)
		case "study":
			embed = HelpStudyEmbed(s.State.User)
		default:
			err = errors.New("옵션을 찾을 수 없습니다.")
		}
	}

	var response *discordgo.InteractionResponse

	if err != nil {
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					ErrorEmbed(err.Error()),
				},
			},
		}
	} else {
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "",
				Flags:   discordgo.MessageFlagsEphemeral,
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							helpSelectMenu,
						},
					},
				},
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		}
	}

	err = s.InteractionRespond(i.Interaction, response)

	// TODO: improve error handling
	if err != nil {
		log.Println(err)
	}
}
