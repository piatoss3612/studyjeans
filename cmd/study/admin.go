package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var adminCmd = discordgo.ApplicationCommand{
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
					Name:  "발표자 등록 마감",
					Value: "close-registration",
				},
				{
					Name:  "발표자 등록 취소",
					Value: "cancel-registration",
				},
				{
					Name:  "발표자료 제출 시작",
					Value: "start-submission",
				},
				{
					Name:  "발표자료 제출 마감",
					Value: "close-submission",
				},
				{
					Name:  "발표 시작",
					Value: "start-presentation",
				},
				{
					Name:  "발표 종료",
					Value: "end-presentation",
				},
				{
					Name:  "피드백 시작",
					Value: "start-feedback",
				},
				{
					Name:  "피드백 종료",
					Value: "end-feedback",
				},
				{
					Name:  "스터디 종료",
					Value: "end-study",
				},
				{
					Name:  "공지 채널 설정",
					Value: "set-notice-channel",
				},
			},
			Required: true,
		},
		{
			Name:        "제목",
			Description: "스터디 제목을 입력해주세요.",
			Type:        discordgo.ApplicationCommandOptionString,
		},
		{
			Name:        "스터디원",
			Description: "스터디원을 선택해주세요.",
			Type:        discordgo.ApplicationCommandOptionUser,
		},
		{
			Name:        "채널",
			Description: "채널을 선택해주세요.",
			Type:        discordgo.ApplicationCommandOptionChannel,
		},
	},
}

func (b *StudyBot) adminHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil || user.ID != b.svc.GetManagerID() {
		return
	}

	options := i.ApplicationCommandData().Options

	cmd := options[0].StringValue()

	var title string
	var u *discordgo.User
	var ch *discordgo.Channel

	for _, o := range options[1:] {
		switch o.Name {
		case "제목":
			title = o.StringValue()
		case "스터디원":
			u = o.UserValue(s)
		case "채널":
			ch = o.ChannelValue(s)
		}
	}

	switch cmd {
	case "create-study":
		b.createStudyHandler(s, i, title)
	case "close-registration":
		b.closeRegistrationHandler(s, i)
	case "cancel-registration":
		b.cancelRegistrationHandler(s, i, u)
	case "start-submission":
		b.startSubmissionHandler(s, i)
	case "close-submission":
		b.closeSubmissionHandler(s, i)
	case "start-presentation":
		b.startPresentationHandler(s, i)
	case "end-presentation":
		b.endPresentationHandler(s, i)
	case "start-feedback":
		b.startFeedbackHandler(s, i)
	case "end-feedback":
		b.endFeedbackHandler(s, i)
	case "end-study":
		b.endStudyHandler(s, i)
	case "set-notice-channel":
		b.setNoticeChannelHandler(s, i, ch)
	default:
		return
	}
}

func (b *StudyBot) createStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate, title string) {
	guildID := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guildID != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	// check if title is empty
	if title == "" {
		// TODO: error handling
		log.Println("empty title")
		return
	}

	// get all members in the guild
	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// get all id of non-bot members
	var memberIDs []string

	for _, m := range members {
		if m.User == nil || m.User.Bot {
			continue
		}

		memberIDs = append(memberIDs, m.User.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// create a study
	err = b.svc.CreateStudy(ctx, i.Member.User.ID, title, memberIDs)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "스터디 생성", fmt.Sprintf("**<%s>**가 생성되었습니다.", title)))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디가 생성되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) closeRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guildID != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// close registration
	err := b.svc.FinishRegistration(ctx, i.Member.User.ID)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자 등록 마감", "발표자 등록이 마감되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자 등록이 마감되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) cancelRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// cancel registration
	err := b.svc.ChangeMemberRegistration(ctx, i.Member.User.ID, u.ID, "", "", false)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("<@%s>님의 발표자 등록이 취소되었습니다.", u.ID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) startSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start submission
	err := b.svc.StartSubmission(ctx, i.Member.User.ID)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자료 제출 시작", "발표자료 제출이 시작되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자료 제출이 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) closeSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// close submission
	err := b.svc.FinishSubmission(ctx, i.Member.User.ID)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표자료 제출 마감", "발표자료 제출이 마감되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표자료 제출이 마감되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) startPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start presentation
	err := b.svc.StartPresentation(ctx, i.Member.User.ID)

	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표 시작", "발표가 시작되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표가 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) endPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// end presentation
	err := b.svc.FinishPresentation(ctx, i.Member.User.ID)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "발표 종료", "발표가 종료되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표가 종료되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) startFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild := b.svc.GetGuildID()

	// check if the command is executed in the correct guild
	if guild != i.GuildID {
		// TODO: error handling
		log.Println("wrong guild")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// start feedback
	err := b.svc.StartReview(ctx, i.Member.User.ID)
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	noticeCh := b.svc.GetNoticeChannelID()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(noticeCh, EmbedTemplate(s.State.User, "피드백 시작", "피드백이 시작되었습니다."))
	if err != nil {
		// TODO: error handling
		log.Println(err)
		return
	}

	// send a response message
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "피드백이 시작되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// TODO: error handling
		log.Println(err)
	}
}

func (b *StudyBot) endFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) endStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) setNoticeChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) {
}
