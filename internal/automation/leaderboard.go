package automation

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
	_ "github.com/mattn/go-sqlite3"
)

type Leaderboard struct {
	session        *discordgo.Session
	db             *sql.DB
	channelID      string
	updateInterval time.Duration
	allowedMaps    []config.MapInfo
	guildID        string
}

func NewLeaderboardUpdater(session *discordgo.Session, db *sql.DB, cfg *config.Config) *Leaderboard {
	if cfg.LeaderboardsChannelID == "" {
		log.Println("[DISCORD] LEADERBOARDS_CHANNEL_ID not set, 'Leaderboards Updater' feature disabled")
		return nil
	}
	return &Leaderboard{
		session:        session,
		db:             db,
		channelID:      cfg.LeaderboardsChannelID,
		updateInterval: cfg.UpdateInterval,
		allowedMaps:    cfg.AllowedMaps,
		guildID:        cfg.DiscordGuildID,
	}
}

func (lb *Leaderboard) Start() {
	if lb == nil {
		return
	}
	log.Println("[DISCORD] Starting 'Leaderboards Updater'...")

	ticker := time.NewTicker(lb.updateInterval)
	go func() {
		lb.updateLeaderboards()
		for range ticker.C {
			lb.updateLeaderboards()
		}
	}()
}

func (lb *Leaderboard) updateLeaderboards() {

	for _, mapInfo := range lb.allowedMaps {
		mapName := mapInfo.MapName
		oldMessage, err := lb.session.ChannelMessage(lb.channelID, mapInfo.MessageID)
		if err != nil {
			log.Printf("[DISCORD] Failed to fetch %s leaderboard message: %v", mapName, err)
			continue
		}
		rows := helpers.LeaderboardReader(lb.db, mapName, lb.allowedMaps)
		newMessage := helpers.TableConstructor(mapName, rows)
		newMessage += fmt.Sprintf("\nhttps://discord.com/channels/%s/%s", lb.guildID, mapInfo.ChannelID)

		if oldMessage.Content == newMessage {
			continue
		} else {
			_, err = lb.session.ChannelMessageEdit(lb.channelID, mapInfo.MessageID, newMessage)
			if err != nil {
				log.Printf("[DISCORD] Failed to update %s leaderboard message: %v", mapName, err)
				continue
			}
		}
	}
}
