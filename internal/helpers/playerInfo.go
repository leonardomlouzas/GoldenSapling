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
		log.Printf("[SECURITY] Attempted to query an invalid table name: %s", mapName)
		return nil
	}

	playerInfo := &PlayerInfo{}
	query := fmt.Sprintf(`
	SELECT MIN(time_score)
	FROM "%s"
	WHERE player_name = "%s"`, mapName, playerName)

	row := db.QueryRow(query)
	var bestTime int
	err := row.Scan(&bestTime)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve best time for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}
	playerInfo.BestTime = ConvertSecondsToTimer(bestTime)

	query = fmt.Sprintf(`
	SELECT COUNT(time_score)
	FROM "%s"
	WHERE player_name = "%s"`, mapName, playerName)
	row = db.QueryRow(query)
	err = row.Scan(&playerInfo.TotalRuns)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve total runs for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}

	query = fmt.Sprintf(`
	SELECT time_score
	FROM "%s"
	WHERE player_name = "%s"
	ORDER BY id DESC
	LIMIT 1`, mapName, playerName)
	row = db.QueryRow(query)
	var lastRun int
	err = row.Scan(&lastRun)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve last run for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}
	playerInfo.LastRun = ConvertSecondsToTimer(lastRun)

	query = fmt.Sprintf(`
	SELECT time_score
	FROM "%s"
	WHERE player_name = "%s"
	ORDER BY id ASC
	LIMIT 1`, mapName, playerName)
	row = db.QueryRow(query)
	var firstRun int
	err = row.Scan(&firstRun)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve last run for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}
	playerInfo.FirstRun = ConvertSecondsToTimer(firstRun)

	query = fmt.Sprintf(`
	SELECT SUM(time_score)
	FROM "%s"
	WHERE player_name = "%s"`, mapName, playerName)
	row = db.QueryRow(query)
	var totalTime int
	err = row.Scan(&totalTime)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve total time for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}
	playerInfo.TotalTime = ConvertSecondsToTimer(totalTime)

	query = fmt.Sprintf(`
	SELECT COUNT(time_score)
	FROM "%s"
	WHERE player_name = "%s" AND time_score = %d`, mapName, playerName, bestTime)
	row = db.QueryRow(query)
	err = row.Scan(&playerInfo.BestTimeAmount)

	query = fmt.Sprintf(`
	SELECT MAX(time_score)
	FROM "%s"
	WHERE player_name = "%s"`, mapName, playerName)

	row = db.QueryRow(query)
	var slowestTime int
	err = row.Scan(&slowestTime)
	if err != nil {
		log.Printf("[DISCORD] Failed to retrieve best time for player %s on map %s: %v", playerName, mapName, err)
		return nil
	}
	playerInfo.SlowestRun = ConvertSecondsToTimer(slowestTime)

	return playerInfo

}
