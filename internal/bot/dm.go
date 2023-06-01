package bot

import "github.com/bwmarrin/discordgo"

func (b *StudyBot) sendDMsToAllMember(s *discordgo.Session, e *discordgo.MessageEmbed, guildID string) {
	// get all members
	members, err := s.GuildMembers(guildID, "", 1000)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "send-dms-to-all-member")
		return
	}

	for i := 1; i <= 10; i++ {
		candidates := make([]*discordgo.Member, 0, 1000)

		for _, member := range members {
			// skip if the member is a bot
			if member.User.Bot {
				continue
			}

			// create a dm channel
			ch, err := s.UserChannelCreate(member.User.ID)
			if err != nil {
				b.sugar.Errorw(err.Error(), "event", "send-dms-to-all-member")
				candidates = append(candidates, member)
				continue
			}

			_, err = s.ChannelMessageSendEmbed(ch.ID, e)
			if err != nil {
				b.sugar.Errorw(err.Error(), "event", "send-dms-to-all-member")
				candidates = append(candidates, member)
			}
		}

		if len(candidates) == 0 {
			b.sugar.Infow("sent dms to all members", "event", "send-dms-to-all-member", "guild_id", guildID)
			break
		}

		members = candidates
	}
}

func (b *StudyBot) sendDMToMember(s *discordgo.Session, u *discordgo.User, e *discordgo.MessageEmbed) {
	ch, err := s.UserChannelCreate(u.ID)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "send-dm-to-member")
		return
	}

	_, err = s.ChannelMessageSendEmbed(ch.ID, e)
	if err != nil {
		b.sugar.Errorw(err.Error(), "event", "send-dm-to-member")
		return
	}
}
