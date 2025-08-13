package automation

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

type Leaderboard struct {
	session              *discordgo.Session
	channelID            string
	updateInterval       time.Duration
	firstMapMessageID    string
	gymMapMessageID      string
	ithurtsMapMessageID  string
	strafeitMapMessageID string
	doorMapMessageID     string
	dbPath               string
	allowedMaps          map[string]string
}

func NewLeaderboardUpdater(s *discordgo.Session, cfg *config.Config) *Leaderboard {
	if cfg.LeaderboardsChannelID == "" {
		log.Println("[DISCORD] LEADERBOARDS_CHANNEL_ID not set, 'Leaderboards Updater' feature disabled")
		return nil
	}
	return &Leaderboard{
		session:              s,
		channelID:            cfg.LeaderboardsChannelID,
		updateInterval:       cfg.UpdateInterval,
		firstMapMessageID:    cfg.FirstMapMessageID,
		gymMapMessageID:      cfg.GymMapMessageID,
		ithurtsMapMessageID:  cfg.IthurtsMapMessageID,
		strafeitMapMessageID: cfg.StrafeitMapMessageID,
		doorMapMessageID:     cfg.DoorMapMessageID,
		dbPath:               cfg.DBPath,
		allowedMaps:          cfg.AllowedMaps,
	}
}

func (lb *Leaderboard) Start() {
	if lb == nil {
		return // Service is disabled
	}
	log.Println("[DISCORD] Starting 'Leaderboards Updater'...")

	ticker := time.NewTicker(lb.updateInterval)
	go func() {
		for range ticker.C {
			lb.updateLeaderboards()
		}
	}()
}

func (lb *Leaderboard) updateLeaderboards() {
	for mapName, mapMessageID := range lb.allowedMaps {
		oldMessage, err := lb.session.ChannelMessage(lb.channelID, mapMessageID)
		if err != nil {
			log.Printf("[DISCORD] Failed to fetch %s leaderboard message: %v", mapName, err)
			continue
		}
		rows := lb.leaderboardReader(mapName)
		newMessage := helpers.TableConstructor(mapName, rows)
		if oldMessage.Content == newMessage {
			continue
		} else {
			_, err = lb.session.ChannelMessageEdit(lb.channelID, mapMessageID, newMessage)
			if err != nil {
				log.Printf("[DISCORD] Failed to update %s leaderboard message: %v", mapName, err)
				continue
			}
		}
	}
}

func (lb *Leaderboard) leaderboardReader(mapName string) []helpers.LeaderboardEntry {
	db, err := sql.Open("sqlite3", lb.dbPath)
	if err != nil {
		log.Printf("[DISCORD] Failed to open database while reading leaderboard: %v", err)
		return nil
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT player_name, MIN(time_score) as best_time
		FROM
		(
			SELECT MAX(id) as id, player_name, time_score
			FROM %s
			GROUP BY player_name, time_score
		)
		GROUP BY player_name
		ORDER BY best_time ASC, id DESC
		LIMIT 10;
		`, mapName)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[DISCORD] Failed to execute query while retrieving Leaderboard: %v", err)
		return nil
	}
	defer rows.Close()

	var entries []helpers.LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry helpers.LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.BestTime); err != nil {
			log.Printf("[DISCORD] Failed to scan row while retrieving Leaderboard: %v", err)
			continue
		}
		entry.Rank = rank
		entries = append(entries, entry)
		rank++
	}
	return entries
}
