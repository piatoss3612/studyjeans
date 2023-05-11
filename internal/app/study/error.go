package study

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrManagerNotFound = errors.New("매니저 정보를 찾을 수 없습니다.")
	ErrNotManager      = errors.New("매니저만 사용할 수 있는 명령어입니다.")
	ErrUserNotFound    = errors.New("사용자 정보를 찾을 수 없습니다.")
	ErrChannelNotFound = errors.New("채널 정보를 찾을 수 없습니다.")
	ErrRequiredArgs    = errors.New("필수 인자가 없습니다.")
	ErrInvalidArgs     = errors.New("인자가 올바르지 않습니다.")
	ErrRoundNotFound   = errors.New("진행중인 스터디 라운드 정보를 찾을 수 없습니다.")
	ErrMemberNotFound  = errors.New("스터디에 등록된 사용자 정보를 찾을 수 없습니다.")
	ErrInvalidCommand  = errors.New("올바르지 않은 명령어입니다.")
	ErrRoundAlreadySet = errors.New("이미 진행중인 스터디 라운드가 있습니다.")
)

func errorInteractionRespond(s *discordgo.Session, i *discordgo.InteractionCreate, err error) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{ErrorEmbed(err.Error())},
		},
	})
}

func ErrorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "오류",
		Description: msg,
		Color:       0xff0000,
	}
}
