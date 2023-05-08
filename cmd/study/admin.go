package main

import "github.com/bwmarrin/discordgo"

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

	switch cmd {
	case "create-study":
		studyName := options[1].StringValue()
		b.createStudyHandler(s, i, studyName)
	case "close-registration":
		b.closeRegistrationHandler(s, i)
	case "cancel-registration":
		u := options[1].UserValue(s)
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
		channel := options[1].ChannelValue(s)
		b.setNoticeChannelHandler(s, i, channel.ID)
	default:
		return
	}
}

func (b *StudyBot) createStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate, studyName string) {

}

func (b *StudyBot) closeRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) cancelRegistrationHandler(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) {
}

func (b *StudyBot) startSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) closeSubmissionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) startPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) endPresentationHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) startFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) endFeedbackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) endStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {}

func (b *StudyBot) setNoticeChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) {
}
