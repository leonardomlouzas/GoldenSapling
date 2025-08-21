package helpers

import "strings"

func MapNameNormalizer(mapName string) string {
	normalized := strings.ToLower(mapName)
	normalized = strings.TrimSpace(normalized)
	normalized = strings.ReplaceAll(normalized, " ", "")
	return normalized
}
