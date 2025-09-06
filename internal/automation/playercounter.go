package automation

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

type PlayerCounter struct {
	session        *discordgo.Session
	channelID      string
	lastKnownName  string
	serverListURL  string
	gamePath       string
	updateInterval time.Duration
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
		session:        s,
		channelID:      cfg.PlayerCountChannelID,
		serverListURL:  cfg.R5RServerListURL,
		updateInterval: cfg.UpdateInterval,
		gamePath:       cfg.GamePath,
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

	req, err := http.NewRequest("POST", pc.serverListURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w while 'Player Counter' execution", err)
	}
	req.Header.Set("User-Agent", "GoldenSaplingBotTreeree/1.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Encoding", "identity")
	req.ContentLength = 0

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w while 'Player Counter' execution", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received non-200 response code: %d while 'Player Counter' execution", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w while 'Player Counter' execution", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %w while 'Player Counter' execution", err)
	}

	servers := result["servers"].([]interface{})
	for _, server := range servers {
		srv := server.(map[string]interface{})
		if srv["name"] == "[NA] MOVEMENT HUB" {
			playerCountStr, ok := srv["playerCount"].(string)
			if !ok {
				return 0, fmt.Errorf("playerCount is not a string while 'Player Counter' execution")
			}
			playersInt, err := strconv.Atoi(playerCountStr)
			if err != nil {
				return 0, fmt.Errorf("failed to convert player count to integer: %w while 'Player Counter' execution", err)
			}
			return playersInt, nil
		}
	}

	err = helpers.RestartHUB(pc.gamePath)
	if err != nil {
		return 0, fmt.Errorf("bot tried to restart HUB server but got: %s", err)
	}
	return 0, fmt.Errorf("bot restarted HUB successfully")
}

/*
Fetches the player count and updates the Discord channel name if it has changed.
*/
func (pc *PlayerCounter) updateChannelName() {
	var channelName string
	count, err := pc.getPlayerCount()
	if err != nil {
		log.Printf("[DISCORD] Failed to get players count while 'Player Counter' execution... %v", err)
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
