package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

type NewRunEntry struct {
	PlayerName string
	TimeScore  string
	MapName    string
}

func NewRunTable(entries []NewRunEntry) string {
	if len(entries) == 0 {
		return "No new runs recorded."
	}
	const maxLineLength = 38
	const usernameMaxLength = 20

	tableTitle := entries[0].MapName + " - NEW RUNS"
	if len(tableTitle) > maxLineLength {
		tableTitle = entries[0].MapName[:27] + " - NEW RUNS"
	}
	tableTitle = strings.ToUpper(tableTitle)
	padding := int((maxLineLength - len(tableTitle)) / 2)
	titleStyleStart := `[2;37m[2;45m[1;37m`
	titleStyleEnd := `[0m[2;37m[2;45m[0m[2;37m[0m`
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
	table += columnStart + " Username                        Time " + columnEnd + "\n"
	for i, entry := range entries {
		timeInt, err := strconv.Atoi(entry.TimeScore)
		if err != nil {
			continue
		}

		if len(entry.PlayerName) > usernameMaxLength {
			entry.PlayerName = entry.PlayerName[:usernameMaxLength-3] + "..."
		}

		if timeInt >= 3600 {
			entry.PlayerName = fmt.Sprintf("%-27s", entry.PlayerName)
		} else {
			entry.PlayerName = fmt.Sprintf("%-30s", entry.PlayerName)
		}

		line := fmt.Sprintf(" %s %s ", entry.PlayerName, ConvertSecondsToTimer(timeInt))
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

func NewRunReader(entry []byte) NewRunEntry {
	line := string(entry)
	parts := strings.Split(line, "!@#$%THISISTHEIRDATA!@#$%:")

	if len(parts) != 3 {
		return NewRunEntry{}
	}

	return NewRunEntry{
		PlayerName: parts[0],
		TimeScore:  parts[1],
		MapName:    parts[2],
	}
}
