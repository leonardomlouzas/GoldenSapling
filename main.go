package main

import (
	"log"

	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	discordbot "github.com/leonardomlouzas/GoldenSapling/internal/discordBot"
)

func main() {
	cfg := config.NewConfig()
	if cfg.DiscordBotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN is not set in the environment.")
	}

	// 3. Create and Run the Bot
	b, err := discordbot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	b.Run()
}
