package helpers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

func AddRunToTable(db *sql.DB, mapName, playerName string, timerInt int, allowedMaps []config.MapInfo) error {
	if !IsValidTable(mapName, allowedMaps) {
		return errors.New("invalid table")
	}
	_, err := db.Exec(
		fmt.Sprintf(`INSERT INTO "%s" (player_name, time_score) VALUES (?, ?)`, mapName),
		playerName, timerInt,
	)
	if err != nil {
		log.Printf("[DISCORD] Failed to insert new run for %s (%d) in %s: %v", playerName, timerInt, mapName, err)

		return errors.New("failed to insert new run")
	}

	return nil
}

func RemoveRunFromTable(db *sql.DB, mapName, playerName string, timerInt int, allowedMaps []config.MapInfo) error {
	if !IsValidTable(mapName, allowedMaps) {
		return errors.New("invalid table")
	}
	_, err := db.Exec(
		fmt.Sprintf(`DELETE FROM "%s" WHERE player_name = ? AND time_score = ?`, mapName), playerName, timerInt)
	if err != nil {
		log.Printf("[DISCORD] Failed to delete run for %s (%d) in %s: %v", playerName, timerInt, mapName, err)

		return errors.New("failed to delete run")
	}

	return nil
}

func DeleteRunsFromTables(db *sql.DB, mapName, playerName string, allowedMaps []config.MapInfo) error {
	if !IsValidTable(mapName, allowedMaps) && mapName != "all" {
		return errors.New("invalid table")
	}

	var mapList []string
	if mapName == "all" {
		for _, maap := range allowedMaps {
			mapList = append(mapList, maap.MapName)
		}
	} else {
		mapList = append(mapList, mapName)
	}

	for _, mapNamee := range mapList {
		_, err := db.Exec(
			fmt.Sprintf(`DELETE FROM "%s" WHERE player_name = ?`, mapNamee), playerName)
		if err != nil {
			log.Printf("[DISCORD] Failed to delete runs for %s in %s: %v", playerName, mapNamee, err)
			continue
		}
	}

	return nil
}

func RenameRunsFromTables(db *sql.DB, playerName, newPlayerName string, allowedMaps []config.MapInfo) error {

	var mapList []string
	for _, maap := range allowedMaps {
		mapList = append(mapList, maap.MapName)
	}
	for _, mapNamee := range mapList {
		_, err := db.Exec(
			fmt.Sprintf(`UPDATE "%s" SET player_name = ? WHERE player_name = ?`, mapNamee), newPlayerName, playerName)
		if err != nil {
			log.Printf("[DISCORD] Failed to rename player runs for %s in %s: %v", playerName, mapNamee, err)
			continue
		}
	}

	return nil
}
