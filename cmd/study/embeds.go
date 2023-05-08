package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func InfoEmbed(u *discordgo.User, title, createdAt, rebootedAt, uptime string) *discordgo.MessageEmbed {
	if u == nil {
		return ErrorEmbed("유저 정보를 읽을 수 없습니다.")
	}

	return &discordgo.MessageEmbed{
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

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
