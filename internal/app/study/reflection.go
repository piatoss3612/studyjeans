package study

import "github.com/bwmarrin/discordgo"

var reflectionCmd = discordgo.ApplicationCommand{
	Name:        "발표회고",
	Description: "발표회고를 작성합니다.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "내용",
			Description: "발표회고 내용을 입력해주세요.",
			Required:    true,
		},
	},
}

func (b *StudyBot) addReflectionCmd() {
	b.hdr.AddCommand(reflectionCmd, b.reflectionCmdHandler)
}

func (b *StudyBot) reflectionCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return nil
	}

	err := cmd(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "reflection")
		_ = errorInteractionRespond(s, i, err)
	}
}
