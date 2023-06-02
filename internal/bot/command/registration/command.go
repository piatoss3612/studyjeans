package registration

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/study"
	"github.com/piatoss3612/my-study-bot/internal/study/service"
	"github.com/piatoss3612/my-study-bot/internal/utils"
	"go.uber.org/zap"
)

type registrationCmd struct {
	svc service.Service

	sugar *zap.SugaredLogger
}

func NewRegistrationCommand(svc service.Service, sugar *zap.SugaredLogger) command.Command {
	return &registrationCmd{
		svc:   svc,
		sugar: sugar,
	}
}

func (rc *registrationCmd) Register(reg command.Registerer) {
	reg.RegisterCommand(registerCmd, rc.register)
	reg.RegisterCommand(changeCmd, rc.showChangeModal)
	reg.RegisterHandler(changeModalCustomID, rc.submitChangeModal)
}

// register as speaker for presentation
func (rc *registrationCmd) register(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	var name, subject string

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "이름":
			name = option.StringValue()
		case "주제":
			subject = option.StringValue()
		}
	}

	if name == "" || subject == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("이름과 발표 주제는 필수 입력 사항입니다"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// register as speaker
	_, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		MemberID:   user.ID,
		MemberName: name,
		Subject:    subject,
	},
		service.RegisterMember, service.ValidateToRegister)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				registrationEmbed(s.State.User, "등록 완료", "발표자 등록이 완료되었습니다."),
			},
		},
	})
}

// show modal to change registration info
func (rc *registrationCmd) showChangeModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get study
	gs, err := rc.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	if gs.OngoingRoundID == "" {
		return study.ErrRoundNotFound
	}

	if !gs.CurrentStage.IsRegistrationOpened() {
		return errors.Join(study.ErrInvalidStage, errors.New("발표자 등록 정보 변경이 가능한 단계가 아닙니다"))
	}

	// get round
	gr, err := rc.svc.GetRound(ctx, gs.OngoingRoundID)
	if err != nil {
		return err
	}

	// get member
	member, ok := gr.GetMember(user.ID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrMemberNotRegistered
	}

	name := member.Name
	subject := member.Subject

	// show modal
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: changeModalCustomID,
			Title:    "피드백 작성",
			Flags:    discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "speaker-name",
							Label:       "발표자 이름",
							Style:       discordgo.TextInputShort,
							Placeholder: "변경할 발표자 이름을 입력해 주세요.",
							Value:       name,
							Required:    true,
							MaxLength:   20,
							MinLength:   1,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "speaker-subject",
							Label:       "발표 주제",
							Style:       discordgo.TextInputShort,
							Placeholder: "변경할 발표 주제를 입력해 주세요.",
							Value:       subject,
							Required:    true,
							MaxLength:   100,
							MinLength:   1,
						},
					},
				},
			},
		},
	})
}

// submit modal to change registration info
func (rc *registrationCmd) submitChangeModal(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	data := i.ModalSubmitData()

	var name, subject string

	for _, c := range data.Components {
		row, ok := c.(*discordgo.ActionsRow)
		if !ok {
			continue
		}

		for _, rc := range row.Components {
			input, ok := rc.(*discordgo.TextInput)
			if !ok {
				continue
			}

			switch input.CustomID {
			case "speaker-name":
				name = input.Value
			case "speaker-subject":
				subject = input.Value
			}
		}
	}

	if name == "" || subject == "" {
		return errors.Join(study.ErrRequiredArgs, errors.New("이름과 발표 주제는 필수 입력 사항입니다"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// update registration
	_, _, err := rc.svc.UpdateRound(ctx, &service.UpdateParams{
		GuildID:    i.GuildID,
		MemberID:   user.ID,
		MemberName: name,
		Subject:    subject,
	}, service.RegisterMember, service.ValidateToChangeRegistration)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				registrationEmbed(s.State.User, "등록 변경 완료", "발표자 등록 정보 변경이 완료되었습니다."),
			},
		},
	})
}
