package discord

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"
)

var (
	RegisteredCommands []*discordgo.ApplicationCommand

	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Ping server",
		},
		{
			Name:        "healthz",
			Description: "Health check",
		},
		{
			Name:        "version",
			Description: "Get server version",
		},
		{
			Name:        "subscribe",
			Description: "Subscribe events",
		},
		{
			Name:        "unsubscribe",
			Description: "Unsubscribe events",
		},
		{
			Name:                     "maintenance",
			Description:              "Toggle server maintenance status\n",
			DefaultMemberPermissions: &defaultMemberPermissions,
			DMPermission:             &dmPermission,
		},
	}
)

func RegisterCommands(s *discordgo.Session) {
	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Fatalw("Cannot create command", v.Name, err)
		}

		RegisteredCommands[i] = cmd
	}
}
