package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
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

const NoticeModalCustomID = "notice"

func (b *StudyBot) addAdminCmd() {
	b.cmd.AddCommand(adminCmd, b.adminHandler)
	b.cpt.AddComponent(NoticeModalCustomID, b.noticeSubmitHandler)
	b.cpt.AddComponent(stageMoveConfirmButton.CustomID, b.stageMoveConfirmHandler)
	b.cpt.AddComponent(stageMoveCancelButton.CustomID, b.stageMoveCancelHandler)
}

func (b *StudyBot) adminHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return
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
		err = b.createStudyHandler(s, i)
	case "notice":
		err = b.noticeHandler(s, i, txt)
	case "refresh-status":
		err = b.refreshStatusHandler(s, i)
	case "create-study-round":
		err = b.createStudyRoundHandler(s, i, txt)
	case "move-round-stage":
		err = b.moveRoundStageHandler(s, i)
	case "confirm-attendance":
		err = b.checkAttendanceHandler(s, i, u)
	case "register-recorded-content":
		err = b.registerRecordedContentHandler(s, i, txt)
	case "set-notice-channel":
		err = b.setNoticeChannelHandler(s, i, ch)
	case "set-reflection-channel":
		err = b.setReflectionChannelHandler(s, i, ch)
	default:
		err = study.ErrInvalidCommand
	}

	if err != nil {
		b.sugar.Errorw(err.Error(), "event", cmd)
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) createStudyHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	gs, err := b.svc.NewStudy(ctx, &service.NewStudyParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
	})
	if err != nil {
		return err
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			CustomID: NoticeModalCustomID,
			Title:    "스터디 생성",
			Flags:    discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				EmbedTemplate(s.State.User, "스터디가 생성되었습니다.", fmt.Sprintf("스터디 ID: %s", gs.ID)),
			},
		},
	})
}

func (b *StudyBot) noticeHandler(s *discordgo.Session, i *discordgo.InteractionCreate, txt string) error {
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
	gs, err := b.svc.GetStudy(ctx, i.GuildID)
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
			CustomID: NoticeModalCustomID,
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

func (b *StudyBot) noticeSubmitHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
		gs, err := b.svc.GetStudy(ctx, i.GuildID)
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
		go b.sendDMsToAllMember(s, embed, i.GuildID)

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "공지를 전송했습니다.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	err := fn(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "notice-modal-submit")
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) refreshStatusHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	gs, err := b.svc.GetStudy(ctx, i.GuildID)
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

func (b *StudyBot) createStudyRoundHandler(s *discordgo.Session, i *discordgo.InteractionCreate, title string) error {
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
	gs, err := b.svc.NewRound(ctx, &service.NewRoundParams{
		GuildID:   i.GuildID,
		ManagerID: manager.ID,
		Title:     title,
		MemberIDs: memberIDs,
	})
	if err != nil {
		return err
	}

	// update game status
	err = s.UpdateGameStatus(0, gs.CurrentStage.String())
	if err != nil {
		return err
	}

	embed := EmbedTemplate(s.State.User, "스터디 라운드 생성", fmt.Sprintf("**<%s>**가 생성되었습니다.", title))

	// send a notice message
	_, err = s.ChannelMessageSendEmbed(gs.NoticeChannelID, embed)
	if err != nil {
		return err
	}

	// send a DM to all members
	go b.sendDMsToAllMember(s, embed, i.GuildID)

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "스터디가 생성되었습니다.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) moveRoundStageHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	gs, err := b.svc.GetStudy(ctx, i.GuildID)
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

func (b *StudyBot) stageMoveConfirmHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
		gs, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
			GuildID:   i.GuildID,
			ManagerID: manager.ID,
		}, service.MoveStage, service.ValidateToCheckManager, service.ValidateToCheckOngoingRound)
		if err != nil {
			return err
		}

		var embed *discordgo.MessageEmbed

		if gs.CurrentStage == study.StageWait && gs.OngoingRoundID == "" {
			embed = EmbedTemplate(s.State.User, "스터디 라운드가 종료", "스터디 라운드가 종료되었습니다. 다음 라운드를 기다려주세요...")

			// TODO: publish round info
		} else {
			embed = EmbedTemplate(s.State.User, gs.CurrentStage.String(), fmt.Sprintf("**<%s>**이(가) 시작되었습니다.", gs.CurrentStage.String()))

		}

		// send a DM to all members
		go b.sendDMsToAllMember(s, embed, i.GuildID)

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

	err := fn(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "stage-move-confirm")
		_ = errorInteractionRespond(s, i, err)
	}
}

func (b *StudyBot) checkAttendanceHandler(s *discordgo.Session, i *discordgo.InteractionCreate, u *discordgo.User) error {
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
	_, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
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
	go b.sendDMToMember(s, u, embed)

	// send a response message
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**<@%s>**님의 발표 출석이 확인되었습니다.", u.Username),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *StudyBot) registerRecordedContentHandler(s *discordgo.Session, i *discordgo.InteractionCreate, contentURL string) error {
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
	gs, err := b.svc.UpdateRound(ctx, &service.UpdateParams{
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
	go b.sendDMsToAllMember(s, embed, i.GuildID)

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

func (b *StudyBot) setNoticeChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
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
	_, err := b.svc.UpdateStudy(ctx, &service.UpdateParams{
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

func (b *StudyBot) setReflectionChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate, ch *discordgo.Channel) error {
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
	_, err := b.svc.UpdateStudy(ctx, &service.UpdateParams{
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

func (b *StudyBot) stageMoveCancelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	fn := func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		// send a response message
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "스터디 라운드 이동이 취소되었습니다.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	err := fn(s, i)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "stage-move-cancel")
		_ = errorInteractionRespond(s, i, err)
	}
}
