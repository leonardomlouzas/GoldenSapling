package commands

import (
	"database/sql"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

func PlayerInfo(db *sql.DB, playerName string, allowedMaps []config.MapInfo) string {
	return ""
}
