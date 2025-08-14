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
	allowedMaps    map[string]string
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
	query := fmt.Sprintf(`
		SELECT player_name, MIN(time_score) as best_time
		FROM
		(
			SELECT MAX(id) as id, player_name, time_score
			FROM "%s"
			GROUP BY player_name, time_score
		)
		GROUP BY player_name
		ORDER BY best_time ASC, id DESC
		LIMIT 10;
		`, mapName)
	rows, err := lb.db.Query(query)
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

// GetPlayerBest retrieves the best time for a specific player on a specific map.
func (lb *Leaderboard) GetPlayerBest(mapName, playerName string) (*helpers.LeaderboardEntry, error) {
	query := fmt.Sprintf(`
		SELECT player_name, MIN(time_score) as best_time
		FROM "%s"
		WHERE player_name = ?
		GROUP BY player_name;
	`, mapName)

	row := lb.db.QueryRow(query, playerName)
	var entry helpers.LeaderboardEntry
	if err := row.Scan(&entry.PlayerName, &entry.BestTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No record found is not an error, just no data
		}
		return nil, fmt.Errorf("failed to scan player best time: %w", err)
	}

	return &entry, nil
}

// GetPlayerTotalRuns retrieves the total number of runs for a specific player on a specific map.
func (lb *Leaderboard) GetPlayerTotalRuns(mapName, playerName string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM "%s" WHERE player_name = ?;`, mapName)
	row := lb.db.QueryRow(query, playerName)

	var totalRuns int
	if err := row.Scan(&totalRuns); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No runs means 0
		}
		return 0, fmt.Errorf("failed to scan player total runs: %w", err)
	}

	return totalRuns, nil
}
