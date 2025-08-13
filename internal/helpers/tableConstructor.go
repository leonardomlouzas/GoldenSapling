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
	tableTitle := tableName + " Leaderboard"
	if len(tableTitle) > 38 {
		tableTitle = tableName[:26] + " Leaderboard"
	}
	padding := int(38 - len(tableTitle)/2)
	titleStyleStart := `[1;2m[1;37m[1;45m[4;37m`
	titleStyleEnd := `[0m[1;37m[1;45m[4;37m[0m[1;37m[1;45m[0m[1;37m[0m[0m`

	columnStart := `[2;45m[2;37m[1;37m`
	columnEnd := `[0m[2;37m[2;45m[0m[2;45m[0m[2;37m[2;45m[0m[2;37m[0m`

	tableEvenStart := `[2;37m[2;47m[2;30m`
	tableEvenEnd := `[0m[2;37m[2;47m[0m[2;37m[0m`
	tableOddStart := `[2;37m[2;45m`
	tableOddEnd := `[0m[2;37m[0m`

	table := "```ansi\n"
	table += titleStyleStart + strings.Repeat(" ", padding) + tableTitle + strings.Repeat(" ", padding) + titleStyleEnd + "\n"
	table += columnStart + " Rank  Username              Best Time " + columnEnd + "\n"
	for i, entry := range entries {
		if i%2 == 0 {
			table += tableEvenStart + fmt.Sprintf(" %3d  %-20s %d ", entry.Rank, entry.PlayerName, entry.BestTime) + tableEvenEnd + "\n"
		} else {
			table += tableOddStart + fmt.Sprintf(" %3d  %-20s %d ", entry.Rank, entry.PlayerName, entry.BestTime) + tableOddEnd + "\n"
		}
	}

	return table + "```"
}
