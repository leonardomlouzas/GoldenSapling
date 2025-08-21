package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type MapInfo struct {
	MessageID string // ID of the map message used in the leaderboard channel
	MapName   string // Name of the map, used for validation
	ChannelID string // ID of the channel where the map is discussed
}

type Config struct {
	DiscordBotToken       string
	DiscordGuildID        string
	BannedWords           string
	PlayerCountChannelID  string
	PlayerCountFile       string
	UpdateInterval        time.Duration
	LeaderboardsChannelID string
	DBPath                string
	AllowedMaps           []MapInfo
	NewRunsChannelID      string
	NewRunsPath           string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	updateIntervalStr := os.Getenv("UPDATE_INTERVAL")
	updateInterval, err := time.ParseDuration(updateIntervalStr)
	if err != nil || updateIntervalStr == "" {
		updateInterval = 2 * time.Minute // Default to 2 minutes if not set or invalid.
	}

	var allowedMaps []MapInfo
	allowedMapsEnv := os.Getenv("ALLOWED_MAPS")
	for _, mapInfo := range strings.Split(allowedMapsEnv, ",") {
		parts := strings.Split(mapInfo, ":")
		// Expects format map_name:message_id:channel_id
		if len(parts) > 2 {
			allowedMaps = append(allowedMaps, MapInfo{
				MapName:   parts[0],
				MessageID: parts[1],
				ChannelID: parts[2],
			})
		}
	}

	return &Config{
		DiscordBotToken:       os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordGuildID:        os.Getenv("DISCORD_GUILD_ID"),
		BannedWords:           os.Getenv("BANNED_WORDS"),
		PlayerCountChannelID:  os.Getenv("PLAYER_COUNT_CHANNEL_ID"),
		PlayerCountFile:       os.Getenv("PLAYER_COUNT_FILE"),
		UpdateInterval:        updateInterval,
		LeaderboardsChannelID: os.Getenv("LEADERBOARDS_CHANNEL_ID"),
		DBPath:                os.Getenv("DB_PATH"),
		AllowedMaps:           allowedMaps,
		NewRunsChannelID:      os.Getenv("NEW_RUNS_CHANNEL_ID"),
		NewRunsPath:           os.Getenv("NEW_RUNS_PATH"),
	}
}
