package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordBotToken      string
	DiscordGuildID       string
	BannedWords          string
	PlayerCountChannelID string
	PlayerCountFile      string
	UpdateInterval       time.Duration
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	updateIntervalStr := os.Getenv("UPDATE_INTERVAL")
	updateInterval, err := time.ParseDuration(updateIntervalStr)
	if err != nil || updateIntervalStr == "" {
		updateInterval = 2 * time.Minute // Default to 5 minutes if not set or invalid.
	}

	return &Config{
		DiscordBotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordGuildID:       os.Getenv("DISCORD_GUILD_ID"),
		BannedWords:          os.Getenv("BANNED_WORDS"),
		PlayerCountChannelID: os.Getenv("PLAYER_COUNT_CHANNEL_ID"),
		PlayerCountFile:      os.Getenv("PLAYER_COUNT_FILE"),
		UpdateInterval:       updateInterval,
	}
}
