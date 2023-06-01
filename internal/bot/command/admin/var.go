package admin

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	adminCmd = discordgo.ApplicationCommand{
		Name:        "매니저",
		Description: "스터디 관리 명령어입니다. 매니저만 사용할 수 있습니다.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "명령어",
				Description: "사용할 명령어를 선택해주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "스터디 생성",
						Value: "create-study",
					},
					{
						Name:  "공지",
						Value: "notice",
					},
					{
						Name:  "상태 갱신",
						Value: "refresh-status",
					},
					{
						Name:  "스터디 라운드 생성",
						Value: "create-study-round",
					},
					{
						Name:  "스터디 라운드 이동",
						Value: "move-round-stage",
					},
					{
						Name:  "발표자 참여 확정",
						Value: "confirm-attendance",
					},
					{
						Name:  "발표 녹화 자료 등록",
						Value: "register-recorded-content",
					},
					{
						Name:  "공지 채널 설정",
						Value: "set-notice-channel",
					},
					{
						Name:  "회고 채널 설정",
						Value: "set-reflection-channel",
					},
					{
						Name:  "스프레드시트 설정",
						Value: "set-spreadsheet",
					},
				},
				Required: true,
			},
			{
				Name:        "텍스트",
				Description: "텍스트를 입력해주세요.",
				Type:        discordgo.ApplicationCommandOptionString,
			},
			{
				Name:        "사용자",
				Description: "사용자를 선택해주세요.",
				Type:        discordgo.ApplicationCommandOptionUser,
			},
			{
				Name:        "채널",
				Description: "채널을 선택해주세요.",
				Type:        discordgo.ApplicationCommandOptionChannel,
			},
		},
	}
	noticeTextInput = discordgo.TextInput{
		CustomID:    "notice",
		Label:       "공지",
		Style:       discordgo.TextInputParagraph,
		Placeholder: "공지 내용을 입력해주세요.",
		Required:    true,
		MaxLength:   3000,
		MinLength:   10,
	}
	stageMoveConfirmButton = discordgo.Button{
		CustomID: "confirm-move-stage",
		Label:    "확인",
		Style:    discordgo.SuccessButton,
	}
	stageMoveCancelButton = discordgo.Button{
		CustomID: "cancel-move-stage",
		Label:    "취소",
		Style:    discordgo.DangerButton,
	}
)

const noticeModalCustomID = "notice"

func EmbedTemplate(u *discordgo.User, title, description string, url ...string) *discordgo.MessageEmbed {
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

	if len(url) > 0 {
		embed.URL = url[0]
	}

	return embed
}
