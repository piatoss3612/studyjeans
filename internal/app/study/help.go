package study

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	helpCmd = discordgo.ApplicationCommand{
		Name:        "도움",
		Description: "도움말을 확인합니다.",
	}
	helpSelectMenu = discordgo.SelectMenu{
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
	helpLinkButton = discordgo.Button{
		Emoji: discordgo.ComponentEmoji{
			Name: "🔥",
		},
		Label: "큰 결심 하기",
		Style: discordgo.LinkButton,
		URL:   "https://github.com/piatoss3612",
	}
)

func (b *StudyBot) addHelpCmd() {
	b.hdr.AddCommand(helpCmd, b.helpCmdHandler)
	b.chdr.AddHandleFunc(helpSelectMenu.CustomID, b.helpSelectMenuHandler)
}

func (b *StudyBot) helpCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{HelpIntroEmbed(s.State.User)},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						helpSelectMenu,
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{helpLinkButton},
				},
			},
		},
	})
	if err != nil {
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) helpSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var embed *discordgo.MessageEmbed

		data := i.MessageComponentData().Values
		if len(data) == 0 {
			return errors.Join(ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
		}

		switch data[0] {
		case "default":
			embed = HelpDefaultEmbed(s.State.User)
		case "study":
			embed = HelpStudyEmbed(s.State.User)
		default:
			return errors.Join(ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
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

func HelpIntroEmbed(u *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       "도움말",
		Description: "아래의 도움말 옵션을 선택해주세요!",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: u.AvatarURL(""),
		},
		Color: 16777215,
	}
}

func HelpDefaultEmbed(u *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       "❔ 기본 명령어",
		Description: "> 명령어 사용 예시: /[명령어]",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "도움말",
				Value: "명령어 도움말 확인",
			},
			{
				Name:  "프로필",
				Value: "발표 진스의 프로필 확인",
			},
		},
	}
}

func HelpStudyEmbed(u *discordgo.User) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       "📚 스터디 명령어",
		Description: "> 명령어 사용 예시: /[명령어]",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "내-정보",
				Value: "내 스터디 등록 정보 확인",
			},
			{
				Name:  "발표자-등록",
				Value: "발표자로 등록",
			},
			{
				Name:  "발표자-등록-취소",
				Value: "발표자 등록 취소",
			},
			{
				Name:  "발표-자료-제출",
				Value: "발표 자료 제출",
			},
			{
				Name:  "피드백",
				Value: "발표자에게 피드백 전송",
			},
		},
	}
}
