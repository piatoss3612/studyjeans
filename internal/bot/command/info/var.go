package info

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/study"
)

var (
	myStudyInfoCmd = discordgo.ApplicationCommand{
		Name:        "내-정보",
		Description: "나의 스터디 라운드 등록 정보를 확인합니다.",
	}
	studyInfoCmd = discordgo.ApplicationCommand{
		Name:        "스터디-정보",
		Description: "스터디 정보를 확인합니다.",
	}
	studyRoundInfoCmd = discordgo.ApplicationCommand{
		Name:        "라운드-정보",
		Description: "진행중인 스터디 라운드 정보를 확인합니다.",
	}
	speakerInfoSelectMenu = discordgo.SelectMenu{
		CustomID:    "speaker-info",
		Placeholder: "발표자 등록 정보 검색 🔍",
		MenuType:    discordgo.UserSelectMenu,
	}
)

func studyInfoEmbed(u *discordgo.User, s *study.Study) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:     "스터디 정보",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: u.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "관리자",
				Value:  fmt.Sprintf("```%s```", s.ManagerID),
				Inline: true,
			},
			{
				Name:  "생성일",
				Value: fmt.Sprintf("```%s```", s.CreatedAt.Format(time.RFC3339)),
			},
			{
				Name:   "총 라운드 수",
				Value:  fmt.Sprintf("```%d```", s.TotalRound),
				Inline: true,
			},
			{
				Name:   "진행 단계",
				Value:  fmt.Sprintf("```%s```", s.CurrentStage),
				Inline: true,
			},
			{
				Name: "이전 라운드 조회",
				Value: fmt.Sprintf("```%s```", func() string {
					if s.SpreadsheetURL == "" {
						return "미등록"
					}
					return s.SpreadsheetURL
				}()),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func studyRoundInfoEmbed(u *discordgo.User, r *study.Round) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    u.Username,
			IconURL: u.AvatarURL(""),
		},
		Title:     "현재 진행중인 스터디 라운드 정보",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: u.AvatarURL("")},
		Fields: []*discordgo.MessageEmbedField{

			{
				Name:   "번호",
				Value:  fmt.Sprintf("```%d```", r.Number),
				Inline: true,
			},
			{
				Name:   "제목",
				Value:  fmt.Sprintf("```%s```", r.Title),
				Inline: true,
			},
			{
				Name:   "진행 단계",
				Value:  fmt.Sprintf("```%s```", r.Stage.String()),
				Inline: true,
			},
			{
				Name: "발표 결과 자료",
				Value: fmt.Sprintf("```%s```", func() string {
					if r.ContentURL == "" {
						return "미등록"
					}
					return r.ContentURL
				}()),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func speakerInfoEmbed(u *discordgo.User, m study.Member) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s님의 발표 정보", u.Username),
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
				Name: "발표 참여",
				Value: func() string {
					if m.Attended {
						return "```O```"
					}
					return "```X```"
				}(),
				Inline: true,
			},
			{
				Name: "발표 주제",
				Value: func() string {
					if m.Subject == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.Subject)
				}(),
			},
			{
				Name: "발표 자료",
				Value: func() string {
					if m.ContentURL == "" {
						return "```미등록```"
					}
					return fmt.Sprintf("```%s```", m.ContentURL)
				}(),
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     16777215,
	}
}

func errorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
