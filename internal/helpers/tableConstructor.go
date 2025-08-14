package helpers

import (
	"fmt"
	"strings"
)

type LeaderboardEntry struct {
	Rank       int
	PlayerName string
	BestTime   int
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
