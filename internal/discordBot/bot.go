package discordbot

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/automation"
	"github.com/leonardomlouzas/GoldenSapling/internal/commands"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
	_ "github.com/mattn/go-sqlite3"
)

type Bot struct {
	Session       *discordgo.Session
	Config        *config.Config
	AutoBan       *automation.AutoBan
	LinkFixer     *automation.LinkFixer
	PlayerCounter *automation.PlayerCounter
	TempMessenger *automation.TempMessenger
	Leaderboarder *automation.Leaderboard
	DB            *sql.DB
}

func (b *Bot) getCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Displays information about the bot and its commands.",
		},
		{
			Name:        "leaderboard",
			Description: "Displays the leaderboard for the current map post.",
		},
		{
			Name:        "personal_best",
			Description: "Displays the personal best for a player on the current map post.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player's nickname.",
					Required:    true,
				},
			},
		},
		{
			Name:        "personal_total_runs",
			Description: "Displays the total amount of runs for a player on the current map.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player's nickname.",
					Required:    true,
				},
			},
		},
	}
}

/*
Generates a hash for command duplication checks
*/
func hashCommand(cmd *discordgo.ApplicationCommand) string {
	comparable := struct {
		Name        string                                `json:"name"`
		Description string                                `json:"description"`
		Options     []*discordgo.ApplicationCommandOption `json:"options,omitempty"`
		Type        discordgo.ApplicationCommandType      `json:"type,omitempty"`
	}{
		Name:        cmd.Name,
		Description: cmd.Description,
		Options:     cmd.Options,
		Type:        cmd.Type,
	}

	if comparable.Type == 0 {
		comparable.Type = discordgo.ChatApplicationCommand
	}

	b, _ := json.Marshal(comparable)
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

/*
Creates a new Discord bot instance with the provided configuration.
*/
func New(cfg *config.Config) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	autoBanService := automation.NewAutoBan(cfg)
	tempMessengerService := automation.NewTempMessenger()
	playerCounterService := automation.NewPlayerCounter(dg, cfg)
	leaderboardService := automation.NewLeaderboardUpdater(dg, db, cfg)
	linkFixerService, err := automation.NewLinkFixer()
	if err != nil {
		return nil, fmt.Errorf("failed to create LinkFixer service: %w", err)
	}

	return &Bot{
		Session:       dg,
		Config:        cfg,
		DB:            db,
		AutoBan:       autoBanService,
		LinkFixer:     linkFixerService,
		PlayerCounter: playerCounterService,
		TempMessenger: tempMessengerService,
		Leaderboarder: leaderboardService,
	}, nil
}

/*
Synchronizes the bot's commands with Discord and cleans up any outdated commands.
It checks for existing commands, updates them if necessary, creates new ones, and deletes any commands that are no longer needed.
*/
func (b *Bot) syncAndCleanCommands() {
	appID := b.Session.State.User.ID
	guildID := b.Config.DiscordGuildID
	commands := b.getCommands()

	log.Println("[DISCORD] Synchronizing commands...")

	existingCommands, err := b.Session.ApplicationCommands(appID, guildID)
	if err != nil {
		log.Printf("[DISCORD] Failed to fetch existing commands: %v", err)
		return
	}

	keep := make(map[string]bool)

	for _, newCmd := range commands {
		found := false
		newHash := hashCommand(newCmd)

		for _, oldCmd := range existingCommands {
			if oldCmd.Name == newCmd.Name {
				found = true
				oldHash := hashCommand(oldCmd)
				keep[oldCmd.ID] = true

				if oldHash != newHash {
					_, err := b.Session.ApplicationCommandEdit(appID, guildID, oldCmd.ID, newCmd)
					if err != nil {
						log.Printf("[DISCORD] Failed to update command %s: %v", oldCmd.Name, err)
					} else {
						log.Printf("[DISCORD] Updated command: %s", oldCmd.Name)
					}
				}
				break
			}
		}
		if !found {
			_, err := b.Session.ApplicationCommandCreate(appID, guildID, newCmd)
			if err != nil {
				log.Printf("[DISCORD] Failed to create command %s: %v", newCmd.Name, err)
			} else {
				log.Printf("[DISCORD] Created command: %s", newCmd.Name)
			}
		}
	}

	for _, oldCmd := range existingCommands {
		if !keep[oldCmd.ID] {
			err := b.Session.ApplicationCommandDelete(appID, guildID, oldCmd.ID)
			if err != nil {
				log.Printf("[DISCORD] Failed to delete command %s: %v", oldCmd.Name, err)
			} else {
				log.Printf("[DISCORD] Deleted command: %s", oldCmd.Name)
			}
		}
	}
}

/*
Runs the Discord bot, setting up event handlers and starting the session.
*/
func (b *Bot) Run() {
	b.Session.AddHandler(b.ready)
	b.Session.AddHandler(b.interactionCreate)
	b.Session.AddHandler(b.AutoBan.MessageCreateHandler)
	b.Session.AddHandler(b.LinkFixer.MessageCreateHandler)
	b.Session.AddHandler(b.TempMessenger.MessageCreateHandler)

	b.Session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	err := b.Session.Open()
	if err != nil {
		log.Fatalf("[DISCORD] Error opening connection: %v", err)
	}

	log.Println("[DISCORD] Starting bot...")

	// Wait for a termination signal (like Ctrl+C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("[DISCORD] Shutting down bot...")
	b.DB.Close()
	b.Session.Close()
}

/*
Handles the bot's ready event, which is triggered when the bot has successfully connected to Discord.
*/
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Managing Movement HUB")
	b.syncAndCleanCommands()
	b.PlayerCounter.Start()
	b.Leaderboarder.Start()
	log.Println("[DISCORD] Bot is ready!")
}

/*
Handles interaction events, such as slash commands.
*/
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "help":
		b.handleHelpCommand(s, i)
	case "leaderboard":
		b.handleLeaderboardCommand(s, i)
	case "personal_best":
		b.handlePersonalBestCommand(s, i)
	case "personal_total_runs":
		b.handlePersonalTotalRunsCommand(s, i)
	}
}

func (b *Bot) handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				commands.HelpDiscord(),
			},
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to help command: %v", err)
	}
}

func (b *Bot) handleLeaderboardCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Printf("[DISCORD] Failed to get channel: %v", err)
		return
	}
	mapName := channel.Name
	mapName = strings.ToLower(mapName)
	mapName = strings.ReplaceAll(mapName, " ", "")
	if !helpers.IsValidTable(mapName, b.Config.AllowedMaps) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command can only be used in movement map channels.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send ephemeral message for wrong channel: %v", err)
		}
		return
	}

	content := commands.LeaderboardByMapName(b.DB, mapName, b.Config.AllowedMaps)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to leaderboard command: %v", err)
		return
	}
}

func (b *Bot) handlePersonalBestCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mapName, ok := b.Config.MapChannels[i.ChannelID]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command can only be used in movement map channels.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send ephemeral message for wrong channel: %v", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	nick := options[0].StringValue()

	entry, err := b.Leaderboarder.GetPlayerBest(mapName, nick)
	if err != nil {
		log.Printf("[DISCORD] Error fetching personal best for %s on map %s: %v", nick, mapName, err)
		// You might want to send an ephemeral error message to the user here as well
		return
	}

	var content string
	if entry == nil {
		content = fmt.Sprintf("No records found for player **%s** on map **%s**.", nick, mapName)
	} else {
		content = fmt.Sprintf("The personal best for **%s** on **%s** is: **%s**", entry.PlayerName, mapName, entry.BestTime)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to personal_best command: %v", err)
	}
}

func (b *Bot) handlePersonalTotalRunsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mapName, ok := b.Config.MapChannels[i.ChannelID]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command can only be used in a designated map channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send ephemeral message for wrong channel: %v", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	nick := options[0].StringValue()

	totalRuns, err := b.Leaderboarder.GetPlayerTotalRuns(mapName, nick)
	if err != nil {
		log.Printf("[DISCORD] Error fetching total runs for %s on map %s: %v", nick, mapName, err)
		return
	}

	content := fmt.Sprintf("Player **%s** has a total of **%d** runs on map **%s**.", nick, totalRuns, mapName)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to personal_total_runs command: %v", err)
	}
}
