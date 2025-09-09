package helpers

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

func LastRunsReader(db *sql.DB, playerName, mapName string, allowedMaps []config.MapInfo) []LeaderboardEntry {
	if !IsValidTable(mapName, allowedMaps) {
		log.Printf("[DISCORD] Attempted to query an invalid table name: %s", mapName)
		return nil
	}

	query := fmt.Sprintf(`
		SELECT player_name, time_score
		FROM %s
		WHERE player_name = ?
		ORDER BY id DESC
		LIMIT 10`, mapName)

	rows, err := db.Query(query, playerName)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve last runs for player %s on map %s: %v", playerName, mapName, err)

		return nil
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.PlayerName, &entry.BestTime); err != nil {
			log.Printf("[DISCORD] Failed to scan row for player %s on map %s: %v", playerName, mapName, err)
			continue
		}
		entry.Rank = rank
		entries = append(entries, entry)
		rank++
	}

	if err := rows.Err(); err != nil {
		log.Printf("[DISCORD] Row iteration error for player %s on map %s: %v", playerName, mapName, err)
	}

	return entries
}
