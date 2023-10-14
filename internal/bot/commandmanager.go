package bot

// type ApplicationCommandManager struct {
// 	s *discordgo.Session
// }

// func NewApplicationCommandManager(s *discordgo.Session) *ApplicationCommandManager {
// 	return &ApplicationCommandManager{s: s}
// }

// func (m *ApplicationCommandManager) CommandCreate(guildID string, cmd *discordgo.ApplicationCommand) error {
// 	_, err := m.s.ApplicationCommandCreate(m.s.State.User.ID, guildID, cmd)
// 	return err
// }

// func (m *ApplicationCommandManager) CommandDelete(guildID string, cmdID string) error {
// 	return m.s.ApplicationCommandDelete(m.s.State.User.ID, guildID, cmdID)
// }

// var _ command.CommandManager = (*ApplicationCommandManager)(nil)
