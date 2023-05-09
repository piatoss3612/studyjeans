package main

import (
	"context"
	"errors"
	"log"
	"time"

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

func (b *StudyBot) feedbackSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var reviewer *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		reviewer = i.Member.User
	}

	if reviewer == nil {
		// TODO: error response
		log.Println("reviewer is nil")
		return
	}

	data := i.ModalSubmitData()

	var revieweeID, feedback string

	for _, c := range data.Components {
		row, ok := c.(*discordgo.ActionsRow)
		if !ok {
			continue
		}

		for _, rc := range row.Components {
			input, ok := rc.(*discordgo.TextInput)
			if !ok {
				continue
			}

			switch input.CustomID {
			case "presentor-id":
				revieweeID = input.Value
			case "feedback":
				feedback = input.Value
			}
		}
	}

	if revieweeID == "" || feedback == "" {
		// TODO: error response
		log.Println("presentorID or feedback is empty")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := b.svc.SetReviewer(ctx, i.GuildID, reviewer.ID, revieweeID)
	if err != nil {
		// TODO: error response
		log.Println(err)
		return
	}

	channel, err := s.UserChannelCreate(revieweeID)
	if err != nil {
		// TODO: error response
		log.Println(err)
		return
	}

	_, err = s.ChannelMessageSend(channel.ID, feedback)
	if err != nil {
		// TODO: error response
		log.Println(err)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "피드백이 전송되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error response
		log.Println(err)
	}
}
