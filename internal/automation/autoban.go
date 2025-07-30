package automation

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

type AutoBan struct {
	bannedWords []string
}

/*
NewAutoBan creates and initializes a new AutoBan service.
It processes the banned words from the configuration once at startup.
*/
func NewAutoBan(cfg *config.Config) *AutoBan {
	bannedWordsStr := cfg.BannedWords
	words := strings.Split(bannedWordsStr, ",")
	processedWords := make([]string, 0, len(words))

	for _, word := range words {
		trimmedWord := strings.TrimSpace(word)
		if trimmedWord != "" {
			processedWords = append(processedWords, strings.ToLower(trimmedWord))
		}
	}

	return &AutoBan{
		bannedWords: processedWords,
	}
}

/*
Checks if the message content contains any of the banned words and bans the author if it does.
*/
func (ab *AutoBan) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	messageContent := strings.ToLower(m.Content)

	for _, word := range ab.bannedWords {
		if strings.Contains(messageContent, word) {
			err := s.GuildBanCreateWithReason(m.GuildID, m.Author.ID, "Automatic ban: Use of forbidden language.", 0)
			if err != nil {
				log.Printf("[DISCORD] Failed to auto-ban user %s (%s): %v", m.Author.Username, m.Author.ID, err)
			} else {
				log.Printf("[DISCORD] User %s (%s) has been banned for using a forbidden word.", m.Author.Username, m.Author.ID)
			}
			return
		}
	}
}
