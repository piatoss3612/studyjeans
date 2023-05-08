package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/study"
)

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
				Name:  "발표-자료-제출",
				Value: "발표 자료 제출",
			},
			{
				Name:  "피드백",
				Value: "피드백 제출",
			},
		},
	}
}

func BotInfoEmbed(u *discordgo.User, title, createdAt, rebootedAt, uptime string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title: title,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "이름",
				Value:  fmt.Sprintf("```%s```", u.Username),
				Inline: true,
			},
			{
				Name:   "생성일",
				Value:  fmt.Sprintf("```%s```", createdAt),
				Inline: true,
			},
			{
				Name:   "재부팅",
				Value:  fmt.Sprintf("```%s```", rebootedAt),
				Inline: true,
			},
			{
				Name:   "업타임",
				Value:  fmt.Sprintf("```%s```", uptime),
				Inline: true,
			},
			{
				Name:   "💻 개발자",
				Value:  fmt.Sprintf("```%s```", "piatoss3612"),
				Inline: true,
			},
			{
				Name:  "📝 소스코드",
				Value: fmt.Sprintf("```%s```", "https://github.com/piatoss3612/presentation-helper-bot"),
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: u.AvatarURL("256"),
		},
		Color: 16777215,
	}
}

func MyStudyInfoEmbed(u *discordgo.User, m study.Member) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "나의 스터디 등록 정보",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: u.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "이름",
				Value: func() string {
					if m.Name == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Name)
				}(),
				Inline: true,
			},
			{
				Name: "발표자 등록",
				Value: func() string {
					if m.Registered {
						return "```O```"
					}
					return "```X```"
				}(),
				Inline: true,
			},
			{
				Name: "발표 완료",
				Value: func() string {
					if m.Attended {
						return "```O```"
					}
					return "```X```"
				}(),
				Inline: true,
			},
			{
				Name: "발표주제",
				Value: func() string {
					if m.Subject == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Subject)
				}(),
			},
			{
				Name: "발표자료",
				Value: func() string {
					if m.ContentURL == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.ContentURL)
				}(),
			},
		},
		Color: 16777215,
	}
}

func EmbedTemplate(u *discordgo.User, title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:       title,
		Description: description,
		Color:       16777215,
	}
}

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
