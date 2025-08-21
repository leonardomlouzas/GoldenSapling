package automation

import (
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

type NewRunners struct {
	updateInterval time.Duration
	session        *discordgo.Session
	channelID      string
	folderPath     string
}

func NewRunnersService(s *discordgo.Session, cfg *config.Config) *NewRunners {
	if cfg.NewRunsChannelID == "" {
		log.Println("[DISCORD] NEW_RUNNERS_CHANNEL_ID not set, 'New Runners' feature disabled")
		return nil
	}
	return &NewRunners{
		session:        s,
		channelID:      cfg.NewRunsChannelID,
		updateInterval: cfg.UpdateInterval,
		folderPath:     cfg.NewRunsPath,
	}
}

func (sc *NewRunners) Start() {
	if sc == nil {
		return // Service is disabled
	}
	log.Println("[DISCORD] Starting 'New Runners'...")

	ticker := time.NewTicker(sc.updateInterval)
	go func() {
		// Perform an initial update on start
		sc.updateNewRunners()
		for range ticker.C {
			sc.updateNewRunners()
		}
	}()
}

func (sc *NewRunners) updateNewRunners() {
	if sc == nil {
		return // Service is disabled
	}

	files, err := os.ReadDir(sc.folderPath)
	if err != nil {
		log.Printf("[DISCORD] Failed to read New Runners directory: %v", err)
		return
	}

	entries := make(map[string][]helpers.NewRunEntry)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := sc.folderPath + "/" + file.Name()
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[DISCORD] Failed to read file %s while in updateNewRunners: %v", file.Name(), err)
			continue
		}

		entry := helpers.NewRunReader(content)
		entries[entry.MapName] = append(entries[entry.MapName], entry)

		err = os.Remove(filePath)
		if err != nil {
			log.Printf("[DISCORD] Failed to delete file %s", err)
			continue
		}
	}

	for mapName, newEntries := range entries {
		if len(newEntries) == 0 {
			continue
		}
		for len(newEntries) > 10 {
			_, err = sc.session.ChannelMessageSend(sc.channelID, helpers.NewRunTable(newEntries[:10]))
			if err != nil {
				log.Printf("[DISCORD] Failed to send new runs for map %s: %v", mapName, err)
				continue
			}
			newEntries = newEntries[10:]
		}

		_, err = sc.session.ChannelMessageSend(sc.channelID, helpers.NewRunTable(newEntries))
		if err != nil {
			log.Printf("[DISCORD] Failed to send new runs for map %s: %v", mapName, err)
			continue
		}
	}
}
