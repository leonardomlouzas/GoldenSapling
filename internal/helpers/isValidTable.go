package helpers

import (
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

func IsValidTable(tableName string, validNames []config.MapInfo) bool {
	for _, maap := range validNames {
		if maap.MapName == tableName {
			return true
		}
	}
	return false
}
