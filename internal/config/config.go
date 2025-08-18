package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordBotToken       string
	DiscordGuildID        string
	BannedWords           string
	PlayerCountChannelID  string
	PlayerCountFile       string
	UpdateInterval        time.Duration
	LeaderboardsChannelID string
	DBPath                string
	AllowedMaps           map[string]string
	MapChannels           map[string]string
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

	allowedMaps := make(map[string]string)
	allowedMapsEnv := os.Getenv("ALLOWED_MAPS")
	for _, pair := range strings.Split(allowedMapsEnv, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		allowedMaps[parts[0]] = parts[1]
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
	}
}
