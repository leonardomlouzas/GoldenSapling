package discordbot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Basic help command",
		},
	}
)

type Bot struct {
	Session *discordgo.Session
	Config  *config.Config
}

func New(cfg *config.Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	return &Bot{
		Session: dg,
		Config:  cfg,
	}, nil
}

func (b *Bot) Run() {
	b.Session.AddHandler(b.ready)
	b.Session.AddHandler(b.interactionCreate)

	b.Session.Identify.Intents = discordgo.IntentsGuilds

	err := b.Session.Open()
	if err != nil {
		log.Fatalf("[DISCORD] Error opening connection: %v", err)
	}

	log.Println("[DISCORD] Bot is now running")

	// Wait for a termination signal (like Ctrl+C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("[DISCORD] Shutting down bot...")
	b.Session.Close()
}

func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("[DISCORD] Bot is ready! Registering commands...")
	s.UpdateGameStatus(0, "Managing Movement HUB")

	// Registering commands to a specific guild is instant.
	// Global registration can take up to an hour.
	_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, b.Config.DiscordGuildID, commands)
	if err != nil {
		log.Fatalf("[DISCORD] Could not register commands: %v", err)
	}
	log.Println("[DISCORD] Commands registered successfully.")
}

func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		if i.ApplicationCommandData().Name == "help" {
			b.handleHelpCommand(s, i)
		}
	}
}

func (b *Bot) handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "GoldenSapling Bot Help",
					Description: "This bot helps manage speedruns and leaderboards for your game server.",
					Color:       0xDAA520, // Goldenrod color
					Fields: []*discordgo.MessageEmbedField{
						{Name: "/help", Value: "Shows this help message.", Inline: true},
						{Name: "/leaderboard", Value: "Displays the current leaderboard.", Inline: true},
						{Name: "/personal_best {nick}", Value: "Shows the personal best for a player.", Inline: true},
					},
					Footer: &discordgo.MessageEmbedFooter{Text: "Happy Speedrunning!"},
				},
			},
		},
	})
}
