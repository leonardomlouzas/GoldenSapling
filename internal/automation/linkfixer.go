package automation

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type fixPattern struct {
	regex             *regexp.Regexp
	replacement       string
	captureGroupIndex int
}

type LinkFixer struct {
	patterns []fixPattern
}

func NewLinkFixer() (*LinkFixer, error) {
	patterns := []struct {
		name              string
		regexStr          string
		replacement       string
		captureGroupIndex int
	}{
		{
			name:              "Twitter/X",
			regexStr:          `https?://(?:www\.)?(twitter|x)\.com/(\w+/status/\d+)`,
			replacement:       "https://fxtwitter.com/%s",
			captureGroupIndex: 2,
		},
		{
			name:              "Reddit",
			regexStr:          `https?://(?:www\.)?reddit\.com(/r/\w+/comments/\w+[^?\s]*)`,
			replacement:       "https://rxddit.com%s",
			captureGroupIndex: 1,
		},
	}

	var compiledPatterns []fixPattern
	for _, p := range patterns {
		regex, err := regexp.Compile(p.regexStr)
		if err != nil {
			return nil, fmt.Errorf("[DISCORD] failed to compile %s link fixer regex: %w", p.name, err)
		}
		compiledPatterns = append(compiledPatterns, fixPattern{
			regex:             regex,
			replacement:       p.replacement,
			captureGroupIndex: p.captureGroupIndex,
		})
	}

	return &LinkFixer{
		patterns: compiledPatterns,
	}, nil
}

/*
Checks messages for links that need fixing and replies with a better version.
*/
func (lf *LinkFixer) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	var fixedLinks []string
	for _, p := range lf.patterns {
		matches := p.regex.FindAllStringSubmatch(m.Content, -1)
		for _, match := range matches {
			if len(match) > p.captureGroupIndex {
				path := match[p.captureGroupIndex]
				fixedLinks = append(fixedLinks, fmt.Sprintf(p.replacement, path))
			}
		}
	}

	if len(fixedLinks) == 0 {
		return
	}

	replyContent := strings.Join(fixedLinks, "\n")
	_, err := s.ChannelMessageSendReply(m.ChannelID, replyContent, m.Reference())
	if err != nil {
		log.Printf("[DISCORD] Failed to send fixed link reply for message %s: %v", m.ID, err)
	}
}
