package study

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *StudyBot) helpSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var embed *discordgo.MessageEmbed

		data := i.MessageComponentData().Values
		if data == nil || len(data) == 0 {
			return errors.Join(ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다."))
		}

		switch data[0] {
		case "default":
			embed = HelpDefaultEmbed(s.State.User)
		case "study":
			embed = HelpStudyEmbed(s.State.User)
		default:
			return errors.Join(ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다."))
		}

		response := &discordgo.InteractionResponse{
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

		return s.InteractionRespond(i.Interaction, response)
	}

	err := fn(s, i)
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) feedbackSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var reviewer *discordgo.User

		if i.Member != nil && i.Member.User != nil {
			reviewer = i.Member.User
		}

		if reviewer == nil {
			return errors.Join(ErrUserNotFound, errors.New("리뷰어 정보를 찾을 수 없습니다."))
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
			return errors.Join(ErrRequiredArgs, errors.New("리뷰 대상자의 아이디 또는 피드백 정보를 찾을 수 없습니다."))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := b.svc.SetReviewer(ctx, i.GuildID, reviewer.ID, revieweeID)
		if err != nil {
			return err
		}

		channel, err := s.UserChannelCreate(revieweeID)
		if err != nil {
			return err
		}

		_, err = s.ChannelMessageSend(channel.ID, feedback)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "피드백이 전송되었습니다.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	err := fn(s, i)
	if err != nil {
		// TODO: error response
		log.Println(err)
	}
}
