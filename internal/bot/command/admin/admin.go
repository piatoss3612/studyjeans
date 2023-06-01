package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"go.uber.org/zap"
)

type adminCommand struct {
	svc service.Service
	pub msgqueue.Publisher

	sugar *zap.SugaredLogger
}

func NewAdminCommand(svc service.Service, pub msgqueue.Publisher, sugar *zap.SugaredLogger) command.Command {
	return &adminCommand{
		svc:   svc,
		pub:   pub,
		sugar: sugar,
	}
}

func (ac *adminCommand) Register(reg command.Registerer) {
	// TODO: register command
	reg.RegisterCommand(adminCmd, ac.adminHandler)
	reg.RegisterHandler(noticeModalCustomID, ac.noticeSubmitHandler)
	reg.RegisterHandler(stageMoveConfirmButton.CustomID, ac.stageMoveConfirmHandler)
	reg.RegisterHandler(stageMoveCancelButton.CustomID, ac.stageMoveCancelHandler)
}

func (ac *adminCommand) adminHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
	}

	options := i.ApplicationCommandData().Options

	cmd := options[0].StringValue()

	var txt string
	var u *discordgo.User
	var ch *discordgo.Channel

	for _, o := range options[1:] {
		switch o.Name {
		case "텍스트":
			txt = o.StringValue()
		case "사용자":
			u = o.UserValue(s)
		case "채널":
			ch = o.ChannelValue(s)
		}
	}

	var err error

	switch cmd {
	case "create-study":
		err = ac.createStudyHandler(s, i)
	case "notice":
		err = ac.noticeHandler(s, i, txt)
	case "refresh-status":
		err = ac.refreshStatusHandler(s, i)
	case "create-study-round":
		err = ac.createStudyRoundHandler(s, i, txt)
	case "move-round-stage":
		err = ac.moveRoundStageHandler(s, i)
	case "confirm-attendance":
		err = ac.checkAttendanceHandler(s, i, u)
	case "register-recorded-content":
		err = ac.registerRecordedContentHandler(s, i, txt)
	case "set-notice-channel":
		err = ac.setNoticeChannelHandler(s, i, ch)
	case "set-reflection-channel":
		err = ac.setReflectionChannelHandler(s, i, ch)
	case "set-spreadsheet":
		err = ac.setSpreadsheetHandler(s, i, txt)
	default:
		err = study.ErrInvalidCommand
	}

	if err != nil {
		return err
	}

	return nil
}

func (ac *adminCommand) createStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return err
	}

	if !(guild.OwnerID == manager.ID) {
		return study.ErrNotManager
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create study
	gs, err := ac.svc.NewStudy(ctx, &service.NewStudyParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
	})
	if err != nil {
		return err
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Title: "스터디 생성",
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				EmbedTemplate(s.State.User, "스터디가 생성되었습니다.", fmt.Sprintf("스터디 ID: %s", gs.ID)),
			},
		},
	})
}

func (ac *adminCommand) noticeHandler(s *discordgo.Session, i *discordgo.InteractionCreate, txt string) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get study
	gs, err := ac.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	// check manager
	if !gs.IsManager(manager.ID) {
		return study.ErrNotManager
	}

	// show notice modal
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: noticeModalCustomID,
			Title:    "공지 입력",
			Flags:    discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{noticeTextInput},
				},
			},
		},
	})
}

func (ac *adminCommand) noticeSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get study
	gs, err := ac.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	// check manager
	if !gs.IsManager(manager.ID) {
		return study.ErrNotManager
	}

	data := i.ModalSubmitData()

	var notice string

	for _, c := range data.Components {
		row, ok := c.(*discordgo.ActionsRow)
		if !ok {
			continue
		}

		for _, c := range row.Components {
			input, ok := c.(*discordgo.TextInput)
			if !ok {
				continue
			}

			notice = input.Value
		}
	}

	if notice == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("공지 내용을 입력해주세요"))
	}

	bot := s.State.User

	// get notice channel
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    bot.Username,
			IconURL: bot.AvatarURL(""),
		},
		Title:       "공지",
		Description: notice,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x00ff00,
	}

	// send notice
	_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
	if err != nil {
		return err
	}

	// send notice DM to all members with confirm button
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "공지를 전송했습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) refreshStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get study
	gs, err := ac.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	// check manager
	if !gs.IsManager(manager.ID) {
		return study.ErrNotManager
	}

	// update game status
	err = s.UpdateGameStatus(0, gs.CurrentStage.String())
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표 진스의 상태가 갱신되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) createStudyRoundHandler(s *discordgo.Session, i *discordgo.InteractionCreate, title string) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	// check if title is empty
	if title == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("스터디 제목은 필수입니다"))
	}

	// get all members in the guild
	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		return err
	}

	// get all id of non-bot members
	var memberIDs []string

	for _, m := range members {
		if m.User == nil || m.User.Bot {
			continue
		}

		memberIDs = append(memberIDs, m.User.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// create a round
	gs, err := ac.svc.NewRound(ctx, &service.NewRoundParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
		Title:     title,
		MemberIDs: memberIDs,
	})
	if err != nil {
		return err
	}

	embed := EmbedTemplate(s.State.User, "스터디 라운드 생성", fmt.Sprintf("**<%s>**가 생성되었습니다.", title))

	go func() {
		evt := &event.StudyEvent{
			T: "study.round-created",
			D: fmt.Sprintf("%s: %s", title, gs.CurrentStage.String()),
			C: time.Now(),
		}

		// publish an event
		go ac.publishEvent(evt)
	}()

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
	if err != nil {
		return err
	}

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// update game status
	err = s.UpdateGameStatus(0, gs.CurrentStage.String())
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디가 생성되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) moveRoundStageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get study
	gs, err := ac.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	// check manager
	if !gs.IsManager(manager.ID) {
		return study.ErrNotManager
	}

	next := gs.CurrentStage.Next()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				EmbedTemplate(s.State.User, "스터디 라운드 진행 단계 변경", fmt.Sprintf("스터디 라운드 진행 단계가 **<%s>**로 변경됩니다. 진행하시겠습니까?", next.String())),
			},
			Flags: discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						stageMoveConfirmButton,
						stageMoveCancelButton,
					},
				},
			},
		},
	})
}

func (ac *adminCommand) stageMoveConfirmHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// move stage
	gs, r, err := ac.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
	}, service.MoveStage, service.ValidateToCheckManager, service.ValidateToCheckOngoingRound)
	if err != nil {
		return err
	}

	var embed *discordgo.MessageEmbed

	if gs.CurrentStage == study.StageWait && gs.OngoingRoundID == "" {
		embed = EmbedTemplate(s.State.User, "스터디 라운드 종료", "스터디 라운드가 종료되었습니다. 다음 라운드를 기대해주세요!")

		// publish round info
		go ac.publishRound("study.round-closed", r)
	} else {
		embed = EmbedTemplate(s.State.User, gs.CurrentStage.String(), fmt.Sprintf("**<%s>**이(가) 시작되었습니다.", gs.CurrentStage.String()))
	}

	go func() {
		evt := &event.StudyEvent{
			T: "study.round-progress",
			D: fmt.Sprintf("%s: %s", r.Title, gs.CurrentStage.String()),
			C: time.Now(),
		}

		// publish an event
		go ac.publishEvent(evt)
	}()

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
	if err != nil {
		return err
	}

	// update game status
	err = s.UpdateGameStatus(0, gs.CurrentStage.String())
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디 라운드가 이동되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) checkAttendanceHandler(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	if u == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check attendance
	_, _, err := ac.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
		MemberID:  u.ID,
	}, service.CheckSpeakerAttendance,
		service.ValidateToCheckManager, service.ValidateToCheckAttendance)
	if err != nil {
		return err
	}

	embed := EmbedTemplate(s.State.User, "발표 출석 확인", fmt.Sprintf("**<@%s>**님의 발표 출석이 확인되었습니다.", u.Username))

	// send a DM to the user
	go ac.sendDMToMember(s, u, embed)

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**<@%s>**님의 발표 출석이 확인되었습니다.", u.Username),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) registerRecordedContentHandler(s *discordgo.Session, i *discordgo.InteractionCreate, contentURL string) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// submit round content
	gs, _, err := ac.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		ManagerID:  manager.ID,
		ContentURL: contentURL,
	}, service.SubmitRoundContent,
		service.ValidateToCheckManager, service.ValidateToSubmitRoundContent)
	if err != nil {
		return err
	}

	embed := EmbedTemplate(s.State.User, "발표 영상 등록", "발표 영상이 등록되었습니다.", contentURL)

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "발표 영상이 등록되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) setNoticeChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	// check if the channel is nil
	if ch == nil {
		return study.ErrChannelNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// set notice channel
	_, err := ac.svc.UpdateStudy(ctx, &service.UpdateParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
		ChannelID: ch.ID,
	}, service.UpdateNoticeChannelID, service.ValidateToCheckManager)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("공지 채널이 %s로 설정되었습니다.", ch.Mention()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) setReflectionChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	// check if the channel is nil
	if ch == nil {
		return study.ErrChannelNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// set notice channel
	_, err := ac.svc.UpdateStudy(ctx, &service.UpdateParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
		ChannelID: ch.ID,
	}, service.UpdateReflectionChannelID, service.ValidateToCheckManager)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("회고 채널이 %s로 설정되었습니다.", ch.Mention()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) stageMoveCancelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디 라운드 이동이 취소되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) setSpreadsheetHandler(s *discordgo.Session, i *discordgo.InteractionCreate, url string) error {
	var manager *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		manager = i.Member.User
	}

	if manager == nil {
		return study.ErrManagerNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// set spreadsheet
	_, err := ac.svc.UpdateStudy(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		ManagerID:  manager.ID,
		ContentURL: url,
	}, service.SetSpreadsheetURL, service.ValidateToCheckManager)
	if err != nil {
		return err
	}

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("스터디 시트가 %s로 설정되었습니다.", url),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (ac *adminCommand) publishRound(topic string, round *study.Round) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := ac.pub.Publish(ctx, topic, round)
		if err != nil {
			ac.sugar.Errorw(err.Error(), "event", "publish round", "topic", topic, "try", i+1)
			continue
		}
		return
	}
}

func (ac *adminCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := ac.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			ac.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}
