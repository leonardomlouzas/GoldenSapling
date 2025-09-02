package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertSecondsToTimer(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, remainingSeconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

func ConvetTimerToSeconds(timer string) int {
	parts := strings.Split(timer, ":")

	if len(parts) != 2 {
		return 0
	}

	minutes, err := strconv.Atoi(parts[0])
	seconds, err2 := strconv.Atoi(parts[1])

	if err != nil || err2 != nil {
		return 0
	}

	return minutes*60 + seconds
}
