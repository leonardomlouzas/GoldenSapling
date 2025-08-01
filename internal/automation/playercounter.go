package automation

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

type PlayerCounter struct {
	session         *discordgo.Session
	channelID       string
	lastKnownName   string
	playerCountFile string
	updateInterval  time.Duration
}

/*
Creates a new PlayerCounter service.
It returns nil if the feature is not configured.
*/
func NewPlayerCounter(s *discordgo.Session, cfg *config.Config) *PlayerCounter {
	if cfg.PlayerCountChannelID == "" {
		log.Println("[DISCORD] PLAYER_COUNT_CHANNEL_ID not set, 'Player Counter' feature disabled")
		return nil
	}
	return &PlayerCounter{
		session:         s,
		channelID:       cfg.PlayerCountChannelID,
		playerCountFile: cfg.PlayerCountFile,
		updateInterval:  cfg.UpdateInterval,
	}
}

/*
Starts the periodic update of the channel name.
*/
func (pc *PlayerCounter) Start() {
	if pc == nil {
		return // Service is disabled
	}
	log.Println("[DISCORD] Starting 'Player Counter'...")

	channel, err := pc.session.Channel(pc.channelID)
	if err == nil {
		pc.lastKnownName = channel.Name
	}

	ticker := time.NewTicker(pc.updateInterval)
	go func() {
		// Perform an initial update on start
		pc.updateChannelName()
		for range ticker.C {
			pc.updateChannelName()
		}
	}()
}

func (pc *PlayerCounter) getPlayerCount() (int, error) {
	file, err := os.Open(pc.playerCountFile)
	if err != nil {
		return 0, fmt.Errorf("failed to open player count file: %w while 'Player Counter' execution", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "players_online" {

			amount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return 0, fmt.Errorf("failed to convert the player amount into a number while 'Player Counter' execution")
			}
			return amount, nil
		}
	}
	return 0, fmt.Errorf("player_online string not found in file while 'Player Counter' execution")
}

/*
Fetches the player count and updates the Discord channel name if it has changed.
*/
func (pc *PlayerCounter) updateChannelName() {
	var channelName string
	count, err := pc.getPlayerCount()
	if err != nil {
		log.Printf("[DISCORD] Failed to get players count while 'Player Counter' execution: %v", err)
		channelName = "HUB is offline"
	} else {
		channelName = fmt.Sprintf("Players online: %d", count)
	}

	if channelName == pc.lastKnownName {
		return
	}

	_, err = pc.session.ChannelEdit(pc.channelID, &discordgo.ChannelEdit{Name: channelName})
	if err != nil {
		log.Printf("[DISCORD] Failed to update hub status voice channel while 'Player Counter' execution: %v", err)
	} else {
		pc.lastKnownName = channelName
	}
}
