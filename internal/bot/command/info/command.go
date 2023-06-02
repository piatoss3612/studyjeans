package info

import (
	"context"
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/cache"
	"github.com/piatoss3612/my-study-bot/internal/study"
	"github.com/piatoss3612/my-study-bot/internal/study/service"
	"github.com/piatoss3612/my-study-bot/internal/utils"
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
	reg.RegisterCommand(myStudyInfoCmd, ic.showMyStudyInfo)
	reg.RegisterCommand(studyInfoCmd, ic.showStudyInfo)
	//
	reg.RegisterCommand(studyRoundInfoCmd, ic.showRoundInfo)
	reg.RegisterHandler(speakerInfoSelectMenu.CustomID, ic.speakerInfoSelectMenuHandler)
}

// show the user's study info
func (ic *infoCommand) showMyStudyInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get the study
	gs, err := ic.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	if gs.OngoingRoundID == "" {
		return study.ErrRoundNotFound
	}

	// get the round
	round, err := ic.svc.GetRound(ctx, gs.OngoingRoundID)
	if err != nil {
		return err
	}

	// get the user's info
	member, ok := round.GetMember(user.ID)
	if !ok {
		return study.ErrMemberNotFound
	}

	// set the round to cache
	go ic.setRoundRetry(round, 5*time.Minute)

	// send response
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

// show the study info
func (ic *infoCommand) showStudyInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// command should be invoked only in guild
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// get the study
	gs, err := ic.svc.GetStudy(ctx, i.GuildID)
	if err != nil {
		return err
	}

	// send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{studyInfoEmbed(s.State.User, gs)},
		},
	})
}

// show the round info
func (ic *infoCommand) showRoundInfo(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// command should be invoked only in guild
	user := utils.GetGuildUserFromInteraction(i)
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

		if gs.OngoingRoundID == "" {
			return study.ErrRoundNotFound
		}

		// get round from database
		round, err = ic.svc.GetRound(ctx, gs.OngoingRoundID)
	}
	if err != nil {
		return err
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
	// command should be invoked only in guild
	user := utils.GetGuildUserFromInteraction(i)
	if user == nil {
		return study.ErrUserNotFound
	}

	// get input data
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

		// get round from database
		round, err = ic.svc.GetRound(ctx, gs.OngoingRoundID)
	}
	if err != nil {
		return err
	}

	var embed *discordgo.MessageEmbed

	member, ok := round.GetMember(selectedUserID)
	if !ok {
		embed = errorEmbed("발표자 정보를 찾을 수 없습니다")
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

	// send response
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
