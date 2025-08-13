package commands

import "github.com/bwmarrin/discordgo"

func TempMessage() *discordgo.Message {
	return &discordgo.Message{
		Content: "This is a temporary message.",
	}
}
