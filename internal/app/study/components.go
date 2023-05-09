package study

import "github.com/bwmarrin/discordgo"

var helpSelectMenu = discordgo.SelectMenu{
	CustomID:    "help",
	Placeholder: "도움말 옵션 💡",
	Options: []discordgo.SelectMenuOption{
		{
			Label: "기본",
			Value: "default",
			Emoji: discordgo.ComponentEmoji{
				Name: "❔",
			},
			Description: "기본 명령어 도움말",
		},
		{
			Label: "스터디",
			Value: "study",
			Emoji: discordgo.ComponentEmoji{
				Name: "📚",
			},
			Description: "스터디 명령어 도움말",
		},
	},
}

var feedbackTextInput = discordgo.TextInput{
	CustomID:    "feedback",
	Label:       "피드백",
	Style:       discordgo.TextInputParagraph,
	Placeholder: "피드백을 입력해주세요.",
	Required:    true,
	MaxLength:   1000,
	MinLength:   10,
}
