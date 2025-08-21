package helpers

import "strings"

type NewRunEntry struct {
	PlayerName string
	timeScore  string
	MapName    string
}

func NewRunTable([]NewRunEntry) string {
	return ""
}

func NewRunReader(entry []byte) NewRunEntry {
	line := string(entry)
	parts := strings.Split(line, "!@#$%THISISTHEIRDATA!@#$%:")

	if len(parts) != 3 {
		return NewRunEntry{}
	}

	return NewRunEntry{
		PlayerName: parts[0],
		timeScore:  parts[1],
		MapName:    parts[2],
	}
}
