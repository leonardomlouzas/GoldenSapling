package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type LeaderboardEntry struct {
	Rank       int
	PlayerName string
	BestTime   int
}

func LeaderboardReader(db *sql.DB, mapName string, allowedMaps map[string]string) []LeaderboardEntry {
	if !IsValidTable(mapName, allowedMaps) {
		log.Printf("[SECURITY] Attempted to query an invalid table name: %s", mapName)
		return nil
	}

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
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[DISCORD] Failed to execute query while retrieving Leaderboard: %v", err)
		return nil
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
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

func TableConstructor(tableName string, entries []LeaderboardEntry) string {
	if entries == nil {
		return ""
	}
	const maxLineLength = 38
	const usernameMaxLength = 20

	tableTitle := strings.ToUpper(tableName) + " LEADERBOARD"
	if len(tableTitle) > maxLineLength {
		tableTitle = strings.ToUpper(tableName[:26]) + " LEADERBOARD"
	}
	padding := int((maxLineLength - len(tableTitle)) / 2)
	titleStyleStart := `[1;2m[1;37m[1;45m`
	titleStyleEnd := `[0m[1;37m[0m[0m`

	columnStart := `[2;45m[2;37m[1;37m`
	columnEnd := `[0m[2;37m[2;45m[0m[2;45m[0m[2;37m[2;45m[0m[2;37m[0m`

	tableEvenStart := `[2;37m[2;47m[2;30m`
	tableEvenEnd := `[0m[2;37m[2;47m[0m[2;37m[0m`
	tableOddStart := `[2;40m[2;37m`
	tableOddEnd := `[0m[2;40m[0m[0;2m[0m`

	titleLine := strings.Repeat(" ", padding) + tableTitle + strings.Repeat(" ", padding)
	if len(titleLine) > maxLineLength {
		titleLine = titleLine[:maxLineLength]
	} else if len(titleLine) < maxLineLength {
		titleLine += strings.Repeat(" ", maxLineLength-len(titleLine))
	}
	table := "```ansi\n"
	table += titleStyleStart + titleLine + titleStyleEnd + "\n"
	table += columnStart + " Rank Username              Best Time " + columnEnd + "\n"

	for i, entry := range entries {

		if len(entry.PlayerName) > usernameMaxLength {
			entry.PlayerName = entry.PlayerName[:usernameMaxLength-3] + "..."
		}
		entry.PlayerName = fmt.Sprintf("%-20s", entry.PlayerName)

		line := fmt.Sprintf(" %3d  %s %10d ", entry.Rank, entry.PlayerName, entry.BestTime)
		if len(line) > maxLineLength {
			line = line[:maxLineLength]
		} else if len(line) < maxLineLength {
			line += strings.Repeat(" ", maxLineLength-len(line))
		}

		if i%2 == 0 {
			table += tableEvenStart + line + tableEvenEnd + "\n"
		} else {
			table += tableOddStart + line + tableOddEnd + "\n"
		}
	}

	return table + "```"
}
