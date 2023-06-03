package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/event"
	"github.com/piatoss3612/my-study-bot/internal/msgqueue"
	"github.com/piatoss3612/my-study-bot/internal/study"
	"github.com/piatoss3612/my-study-bot/internal/study/service"
	"github.com/piatoss3612/my-study-bot/internal/utils"
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
	reg.RegisterCommand(adminCmd, ac.adminHandler)
	reg.RegisterHandler(noticeModalCustomID, ac.sendNotice)
	reg.RegisterHandler(stageMoveConfirmButton.CustomID, ac.moveRoundStageConfirm)
}

// handle admin command
func (ac *adminCommand) adminHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// command should be executed in guild
	manager := utils.GetGuildUserFromInteraction(i)
	if manager == nil {
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
		err = ac.createStudy(s, i)
	case "notice":
		err = ac.writeNotice(s, i, txt)
	case "refresh-status":
		err = ac.refreshBotStatus(s, i)
	case "create-study-round":
		err = ac.createRound(s, i, txt)
	case "move-round-stage":
		err = ac.moveRoundStage(s, i)
	case "confirm-attendance":
		err = ac.checkAttendance(s, i, u)
	case "register-recorded-content":
		err = ac.registerRecordedContent(s, i, txt)
	case "set-notice-channel":
		err = ac.setNoticeChannel(s, i, ch)
	case "set-reflection-channel":
		err = ac.setReflectionChannel(s, i, ch)
	case "set-spreadsheet":
		err = ac.setSpreadsheet(s, i, txt)
	default:
		err = study.ErrInvalidCommand
	}
	return err
}

// create study of guild
func (ac *adminCommand) createStudy(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	manager := utils.GetGuildUserFromInteraction(i)
	if manager == nil {
		return study.ErrManagerNotFound
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return err
	}

	// check manager is owner of guild
	if guild.OwnerID != manager.ID {
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

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Title: "스터디 생성",
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				adminEmbed(s.State.User, "스터디가 생성되었습니다.", fmt.Sprintf("스터디 ID: %s", gs.ID)),
			},
		},
	})
}

// show modal for write notice
func (ac *adminCommand) writeNotice(s *discordgo.Session, i *discordgo.InteractionCreate, txt string) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

// send notice to notice channel of guild
func (ac *adminCommand) sendNotice(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

	var content string

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

			content = input.Value
		}
	}

	// check content
	if content == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("공지로 전송할 내용을 입력해주세요"))
	}

	bot := s.State.User
	embed := adminEmbed(bot, "공지", content)

	// send notice DM to all members with confirm button
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// check notice channel and send notice
	if gs.NoticeChannelID != "" {
		_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
		if err != nil {
			return err
		}
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "공지를 전송했습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// refresh bot status
func (ac *adminCommand) refreshBotStatus(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

// create round of study
func (ac *adminCommand) createRound(s *discordgo.Session, i *discordgo.InteractionCreate, title string) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

	embed := adminEmbed(s.State.User, "스터디 라운드 생성", fmt.Sprintf("**<%s>**가 생성되었습니다.", title))

	go func() {
		evt := &event.StudyEvent{
			T: "study.round-created",
			D: fmt.Sprintf("%s: %s", title, gs.CurrentStage.String()),
			C: time.Now(),
		}

		// publish an event
		go ac.publishRoundProgress(evt)
	}()

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// check notice channel and send notice
	if gs.NoticeChannelID != "" {
		_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
		if err != nil {
			return err
		}
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
			Content: "스터디가 생성되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// move round stage
func (ac *adminCommand) moveRoundStage(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

	if gs.CurrentStage.IsNone() || gs.CurrentStage.IsWait() || gs.OngoingRoundID == "" {
		return study.ErrRoundNotFound
	}

	next := gs.CurrentStage.Next()
	embed := adminEmbed(s.State.User, "스터디 라운드 진행 단계 변경",
		fmt.Sprintf("스터디 라운드 진행 단계가 **<%s>**로 변경됩니다. 진행하시겠습니까?", next.String()), 16777215)

	// send a response with confirm button
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						stageMoveConfirmButton,
					},
				},
			},
		},
	})
}

// confirm to move round stage
func (ac *adminCommand) moveRoundStageConfirm(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

	// check if the round is closed
	if gs.CurrentStage == study.StageWait && gs.OngoingRoundID == "" {
		embed = adminEmbed(s.State.User, "스터디 라운드 종료", "스터디 라운드가 종료되었습니다. 다음 라운드를 기대해주세요!")

		// publish round info
		go ac.publishRoundOnRoundClosed(r)
	} else {
		embed = adminEmbed(s.State.User, gs.CurrentStage.String(), fmt.Sprintf("**<%s>**이(가) 시작되었습니다.", gs.CurrentStage.String()))
	}

	go func() {
		evt := &event.StudyEvent{
			T: "study.round-progress",
			D: fmt.Sprintf("%s: %s", r.Title, gs.CurrentStage.String()),
			C: time.Now(),
		}

		// publish an event
		go ac.publishRoundProgress(evt)
	}()

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	// send a notice message
	if gs.NoticeChannelID != "" {
		_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
		if err != nil {
			return err
		}
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

// check attendance
func (ac *adminCommand) checkAttendance(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) error {
	if u == nil {
		return study.ErrUserNotFound
	}

	manager := utils.GetGuildUserFromInteraction(i)
	if manager == nil {
		return study.ErrManagerNotFound
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

	embed := adminEmbed(s.State.User, "발표 출석 확인", fmt.Sprintf("**<@%s>**님의 발표 출석이 확인되었습니다.", u.Username))

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

// register recorded content
func (ac *adminCommand) registerRecordedContent(s *discordgo.Session, i *discordgo.InteractionCreate, contentURL string) error {
	manager := utils.GetGuildUserFromInteraction(i)
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

	embed := adminEmbed(s.State.User, "발표 영상 등록", "발표 영상이 등록되었습니다.")
	embed.URL = contentURL

	// send a DM to all members
	go ac.sendDMsToAllMember(s, embed, i.GuildID)

	if gs.NoticeChannelID != "" {
		// send a notice message
		_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
		if err != nil {
			return err
		}
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

// set notice channel
func (ac *adminCommand) setNoticeChannel(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
	// check if the channel is nil
	if ch == nil {
		return study.ErrChannelNotFound
	}

	manager := utils.GetGuildUserFromInteraction(i)
	if manager == nil {
		return study.ErrManagerNotFound
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

// set reflection channel
func (ac *adminCommand) setReflectionChannel(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
	// check if the channel is nil
	if ch == nil {
		return study.ErrChannelNotFound
	}

	manager := utils.GetGuildUserFromInteraction(i)
	if manager == nil {
		return study.ErrManagerNotFound
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

// set spreadsheet
func (ac *adminCommand) setSpreadsheet(s *discordgo.Session, i *discordgo.InteractionCreate, url string) error {
	manager := utils.GetGuildUserFromInteraction(i)
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
