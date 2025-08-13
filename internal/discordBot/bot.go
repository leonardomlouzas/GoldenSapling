package discordbot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/automation"
	"github.com/leonardomlouzas/GoldenSapling/internal/commands"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
)

type Bot struct {
	Session       *discordgo.Session
	Config        *config.Config
	AutoBan       *automation.AutoBan
	LinkFixer     *automation.LinkFixer
	PlayerCounter *automation.PlayerCounter
	TempMessenger *automation.TempMessenger
}

func (b *Bot) getCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Displays information about the bot and its commands.",
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

	playerCounterService := automation.NewPlayerCounter(dg, cfg)
	autoBanService := automation.NewAutoBan(cfg)
	linkFixerService, err := automation.NewLinkFixer()
	if err != nil {
		return nil, fmt.Errorf("failed to create LinkFixer service: %w", err)
	}
	tempMessengerService := automation.NewTempMessenger()

	return &Bot{
		Session:       dg,
		Config:        cfg,
		AutoBan:       autoBanService,
		LinkFixer:     linkFixerService,
		PlayerCounter: playerCounterService,
		TempMessenger: tempMessengerService,
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
	b.Session.Close()
}

/*
Handles the bot's ready event, which is triggered when the bot has successfully connected to Discord.
*/
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Managing Movement HUB")
	b.syncAndCleanCommands()
	b.PlayerCounter.Start()
	log.Println("[DISCORD] Bot is ready!")
}

/*
Handles interaction events, such as slash commands.
*/
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "help":
		b.handleHelpCommand(s, i)
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

// TODO: CHECK IF UPDATED COMMAND MMESSAGE IS STILL SHOWING UP AND MAKE LEADERBOARD AUTOMATION WORK
