package feedback

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	cmd = discordgo.ApplicationCommand{
		Name:        "피드백",
		Description: "발표자에게 피드백을 보냅니다.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "발표자",
				Description: "피드백을 받을 발표자를 선택해주세요.",
				Type:        discordgo.ApplicationCommandOptionUser,
				Required:    true,
			},
		},
	}
	textInput = discordgo.TextInput{
		CustomID:    "feedback",
		Label:       "피드백",
		Style:       discordgo.TextInputParagraph,
		Placeholder: "피드백을 입력해주세요.",
		Required:    true,
		MaxLength:   1000,
		MinLength:   10,
	}

	feedbackModalCustomID = "feedback-modal"
)

func feedbackEmbed(u *discordgo.User, content string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "익명",
			IconURL: u.AvatarURL(""),
		},
		Title:       "피드백",
		Description: content,
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}
