package commands

import (
	"database/sql"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func LeaderboardByMapName(db *sql.DB, mapName string, allowedMaps []config.MapInfo) string {
	entries := helpers.LeaderboardReader(db, mapName, allowedMaps)
	if entries == nil {
		return "An error occurred while fetching leaderboard data."
	}

	if len(entries) == 0 {
		return "No records found for this map yet."
	}

	return helpers.TableConstructor(mapName, entries)
}
