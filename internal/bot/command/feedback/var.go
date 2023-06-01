package feedback

import "github.com/bwmarrin/discordgo"

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

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
