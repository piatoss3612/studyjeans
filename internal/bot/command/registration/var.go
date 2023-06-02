package registration

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	registerCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록",
		Description: "발표자 정보를 등록합니다.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "이름",
				Description: "발표자의 이름을 입력해 주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "주제",
				Description: "발표 주제를 입력해 주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	changeCmd = discordgo.ApplicationCommand{
		Name:        "발표자-등록-정보-변경",
		Description: "발표자 등록 정보를 변경합니다.",
	}

	changeModalCustomID = "registration-change-modal"
)

func registrationEmbed(u *discordgo.User, title, description string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       title,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       16777215,
	}

	return embed
}
