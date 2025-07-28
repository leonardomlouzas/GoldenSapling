package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordBotToken string
	DiscordGuildID  string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
	return &Config{
		DiscordBotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordGuildID:  os.Getenv("DISCORD_GUILD_ID"),
	}
}
