package automation

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/commands"
)

// TempMessenger handles sending temporary messages based on chat commands.
type TempMessenger struct{}

// NewTempMessenger creates a new TempMessenger service.
func NewTempMessenger() *TempMessenger {
	return &TempMessenger{}
}

// MessageCreateHandler checks for the `!temp` command and sends a message.
func (tm *TempMessenger) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || m.Content != ">temp" {
		return
	}

	msg := commands.TempMessage()
	if _, err := s.ChannelMessageSend(m.ChannelID, msg.Content); err != nil {
		log.Printf("[DISCORD] Failed to send temp message in channel %s: %v", m.ChannelID, err)
	}
}
