package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

func UpdateTop10File(filepath string, db *sql.DB, allowedMaps []config.MapInfo) {

	top3 := retrieveTop3(db, allowedMaps)
	top10 := retrieveTop10(db, allowedMaps)

	top3names := ``
	ingameLB := ``
	for name, maap := range top3 {
		for _, player := range maap {
			top3names += fmt.Sprintf(`"%s",`, player.PlayerName)

			switch name {
			case "firstmap":
				switch player.Rank {
				case 1:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#1\", < 0, -879, 40643.5 >, < 0, -90, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))

				case 2:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#2\", < 0, -879, 40593.5 >, < 0, -90, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 3:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#3\", < 0, -879, 40543.5 >, < 0, -90, 0 >, false, 1 )\n\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				}
			case "gymmap":
				switch player.Rank {
				case 1:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#1\", < 879, 0, 40643.5 >, < 0, 0, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 2:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#2\", < 879, 0, 40593.5 >, < 0, 0, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 3:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#3\", < 879, 0, 40543.5 >, < 0, 0, 0 >, false, 1 )\n\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				}
			case "ithurtsmap":
				switch player.Rank {
				case 1:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#1\", < 0, 879, 40643.5 >, < 0, 90, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 2:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#2\", < 0, 879, 40593.5 >, < 0, 90, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 3:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#3\", < 0, 879, 40543.5 >, < 0, 90, 0 >, false, 1 )\n\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				}
			case "strafeitmap":
				switch player.Rank {
				case 1:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#1\", < -879, 0, 40643.5 >, < 0, -180, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 2:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#2\", < -879, 0, 40593.5 >, < 0, -180, 0 >, false, 1 )\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				case 3:
					ingameLB += fmt.Sprintf("CreatePanelText(player, \"%s - %s\", \"#3\", < -879, 0, 40543.5 >, < 0, -180, 0 >, false, 1 )\n\n", player.PlayerName, ConvertSecondsToTimer(player.BestTime))
				}
			}
		}
	}
	if len(top3names) > 0 {
		top3names = top3names[:len(top3names)-1]
	}

	top10names := ``
	for _, maap := range top10 {
		for _, player := range maap {
			top10names += fmt.Sprintf(`"%s",`, player.PlayerName)
		}
	}
	if len(top10names) > 0 {
		top10names = top10names[:len(top10names)-1]
	}

	lines := "untyped\n\n"
	lines += "globalize_all_functions\n\n"
	lines += fmt.Sprintf("global const array <string> top3_players = [%s]\n", top3names)
	lines += fmt.Sprintf("global const array <string> top10_players = [%s]\n\n", top10names)
	lines += "void function MH_Spawn_Leaderboards(entity player) {\n"
	lines += ingameLB + "}\n"

	os.WriteFile(filepath, []byte(lines), 0777)
}

func retrieveTop3(db *sql.DB, allowedMaps []config.MapInfo) map[string][]LeaderboardEntry {
	top3Players := make(map[string][]LeaderboardEntry)
	for _, mapInfo := range allowedMaps {
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
		LIMIT 3;
		`, mapInfo.MapName)
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("[DISCORD] Failed to execute query while retrieving top 3: %v", err)
			continue
		}

		var players []LeaderboardEntry
		rank := 1
		for rows.Next() {
			var entry LeaderboardEntry
			if err := rows.Scan(&entry.PlayerName, &entry.BestTime); err != nil {
				log.Printf("[DISCORD] Failed to scan player name: %v", err)
				continue
			}
			entry.Rank = rank
			players = append(players, entry)
			rank++
		}
		rows.Close()
		top3Players[mapInfo.MapName] = players
	}
	return top3Players
}

func retrieveTop10(db *sql.DB, allowedMaps []config.MapInfo) map[string][]LeaderboardEntry {
	top10Players := make(map[string][]LeaderboardEntry)
	for _, mapInfo := range allowedMaps {
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
		LIMIT 7 OFFSET 3;
		`, mapInfo.MapName)
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("[HELPER] Failed to execute query while retrieving top 3: %v", err)
			continue
		}

		var players []LeaderboardEntry
		rank := 1

		for rows.Next() {
			var entry LeaderboardEntry
			if err := rows.Scan(&entry.PlayerName, &entry.BestTime); err != nil {
				log.Printf("[HELPER] Failed to scan player name: %v", err)
				continue
			}
			entry.Rank = rank
			players = append(players, entry)
			rank++
		}
		rows.Close()
		top10Players[mapInfo.MapName] = players
	}
	return top10Players
}
