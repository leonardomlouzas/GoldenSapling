package commands

import (
	"database/sql"
	"fmt"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func LastRuns(db *sql.DB, playerName, mapName string, allowedMaps []config.MapInfo) string {
	entry := helpers.LastRunsReader(db, playerName, mapName, allowedMaps)
	if entry == nil {
		return fmt.Sprintf("No records found for this player on %s", mapName)
	}

	return helpers.TableConstructor(mapName, entry)
}
