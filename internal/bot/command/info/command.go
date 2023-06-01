package info

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/cache"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
	"go.uber.org/zap"
)

type infoCommand struct {
	svc   service.Service
	cache cache.Cache

	sugar *zap.SugaredLogger
}

func NewInfoCommand(svc service.Service, cache cache.Cache, sugar *zap.SugaredLogger) command.Command {
	return &infoCommand{
		svc:   svc,
		cache: cache,
		sugar: sugar,
	}
}

func (ic *infoCommand) Register(reg command.Registerer) {
	reg.RegisterCommand(myStudyInfoCmd, ic.myStudyInfoCmdHandler)
	reg.RegisterCommand(studyInfoCmd, ic.studyInfoCmdHandler)
	reg.RegisterCommand(studyRoundInfoCmd, ic.studyRoundInfoCmdHandler)
	reg.RegisterHandler(speakerInfoSelectMenu.CustomID, ic.speakerInfoSelectMenuHandler)
}

func (ic *infoCommand) myStudyInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gs, err := ic.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	if gs == nil {
		return study.ErrStudyNotFound
	}

	if gs.OngoingRoundID == "" {
		return study.ErrRoundNotFound
	}

	round, err := ic.svc.GetRound(ctx, gs.OngoingRoundID)
	if err != nil {
		return err
	}

	if round == nil {
		return study.ErrRoundNotFound
	}

	member, ok := round.GetMember(user.ID)
	if !ok {
		return study.ErrMemberNotFound
	}

	go ic.setRoundRetry(round, 5*time.Minute)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: user.Mention(),
			Flags:   discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				speakerInfoEmbed(user, member),
			},
		},
	})
}

func (ic *infoCommand) studyInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gs, err := ic.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	if gs == nil {
		return study.ErrStudyNotFound
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{studyInfoEmbed(s.State.User, gs)},
		},
	})
}

func (ic *infoCommand) studyRoundInfoCmdHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	// command should be invoked only in guild
	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var gs *study.Study
	var round *study.Round
	var err error

	exists := ic.roundExists(ctx, i.GuildID)

	// check if round exists in cache
	if exists {
		// get round from cache
		round, err = ic.getRound(ctx, i.GuildID)
	} else {
		gs, err = ic.svc.GetStudy(ctx, i.GuildID)
		if err != nil {
			return err
		}

		if gs == nil {
			return study.ErrStudyNotFound
		}

		if gs.OngoingRoundID == "" {
			return study.ErrRoundNotFound
		}

		// get round from database
		round, err = ic.svc.GetRound(ctx, gs.OngoingRoundID)
	}
	if err != nil {
		return err
	}

	// if round does not exist, return error
	if round == nil {
		return study.ErrRoundNotFound
	}

	// round info embed
	embed := studyRoundInfoEmbed(s.State.User, round)

	// if round does not exist in cache, set round to cache
	if !exists {
		go ic.setRoundRetry(round, 5*time.Second)
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						speakerInfoSelectMenu,
					},
				},
			},
		},
	})
}

func (ic *infoCommand) speakerInfoSelectMenuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var user *discordgo.User

	// command should be invoked only in guild
	if i.Member != nil && i.Member.User != nil {
		user = i.Member.User
	}

	if user == nil {
		return study.ErrUserNotFound
	}

	// get data
	data := i.MessageComponentData().Values
	if len(data) == 0 {
		return errors.Join(study.ErrRequiredArgs, errors.New("옵션을 찾을 수 없습니다"))
	}

	selectedUserID := data[0]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var gs *study.Study
	var round *study.Round
	var err error

	exists := ic.roundExists(ctx, i.GuildID)

	// check if round exists in cache
	if exists {
		// get round from cache
		round, err = ic.getRound(ctx, i.GuildID)
	} else {
		gs, err = ic.svc.GetStudy(ctx, i.GuildID)
		if err != nil {
			return err
		}

		if gs == nil {
			return study.ErrStudyNotFound
		}

		// get round from database
		round, err = ic.svc.GetRound(ctx, gs.OngoingRoundID)
	}
	if err != nil {
		return err
	}

	if round == nil {
		return study.ErrRoundNotFound
	}

	var embed *discordgo.MessageEmbed

	member, ok := round.GetMember(selectedUserID)
	if !ok {
		embed = ErrorEmbed("발표자 정보를 찾을 수 없습니다")
	} else {
		selectedUser, err := s.User(selectedUserID)
		if err != nil {
			return err
		}

		embed = speakerInfoEmbed(selectedUser, member)
	}

	// if round does not exist in cache, set round to cache
	if !exists {
		go ic.setRoundRetry(round, 5*time.Second)
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						speakerInfoSelectMenu,
					},
				},
			},
		},
	})
}
