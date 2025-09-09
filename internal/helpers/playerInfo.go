package helpers

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

type PlayerInfo struct {
	BestTime       string
	TotalRuns      int
	LastRun        string
	FirstRun       string
	TotalTime      string
	BestTimeAmount string
	SlowestRun     string
}

func PlayerInfoReader(db *sql.DB, playerName string, mapName string, allowedMaps []config.MapInfo) *PlayerInfo {
	if !IsValidTable(mapName, allowedMaps) {
		log.Printf("[DISCORD] Attempted to query an invalid table name: %s", mapName)
		return nil
	}

	query := fmt.Sprintf(`
		SELECT
			MIN(time_score),
			COUNT(time_score),
			SUM(time_score),
			MAX(time_score),
			(SELECT time_score FROM "%s" WHERE player_name = ? ORDER BY id ASC LIMIT 1),
			(SELECT time_score FROM "%s" WHERE player_name = ? ORDER BY id DESC LIMIT 1)
		FROM "%s"
		WHERE player_name = ?`, mapName, mapName, mapName)

	var bestTime, totalRuns, totalTime, slowestTime, firstRun, lastRun sql.NullInt64
	err := db.QueryRow(query, playerName, playerName, playerName).Scan(
		&bestTime,
		&totalRuns,
		&totalTime,
		&slowestTime,
		&firstRun,
		&lastRun,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &PlayerInfo{}
		}
		log.Printf("[DISCORD] Failed to retrieve player info for %s on map %s: %v", playerName, mapName, err)
		return nil
	}

	var bestTimeAmount string
	if bestTime.Valid {
		countQuery := fmt.Sprintf(`SELECT COUNT(time_score) FROM "%s" WHERE player_name = ? AND time_score = ?`, mapName)
		err = db.QueryRow(countQuery, playerName, bestTime.Int64).Scan(&bestTimeAmount)
		if err != nil {
			log.Printf("[DISCORD] Failed to retrieve best time amount for player %s on map %s: %v", playerName, mapName, err)
			return nil
		}
	}

	return &PlayerInfo{
		BestTime:       ConvertSecondsToTimer(int(bestTime.Int64)),
		TotalRuns:      int(totalRuns.Int64),
		LastRun:        ConvertSecondsToTimer(int(lastRun.Int64)),
		FirstRun:       ConvertSecondsToTimer(int(firstRun.Int64)),
		TotalTime:      ConvertSecondsToTimer(int(totalTime.Int64)),
		BestTimeAmount: bestTimeAmount,
		SlowestRun:     ConvertSecondsToTimer(int(slowestTime.Int64)),
	}
}
