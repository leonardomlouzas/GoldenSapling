package discordbot

import (
	"crypto/sha256"
	"database/sql"
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
	NewRunners    *automation.NewRunners
	FileUpdater   *automation.FileUpdater
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
			Name:        "player_info",
			Description: "Displays player information for the current map post.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player nickname.",
					Required:    true,
				},
			},
		},
		{
			Name:        "last_runs",
			Description: "Displays the last 10 runs of a player on the current map.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player nickname",
					Required:    true,
				},
			},
		},
		{
			Name:        "zadd",
			Description: "[ADMIN ONLY] Manually add a new run",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player nickname (case sensitive)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timer",
					Description: "The player run timer (00:00)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "map_name",
					Description: "The player run map",
					Required:    true,
				},
			},
		},
		{
			Name:        "zremove",
			Description: "[ADMIN ONLY] Manually remove a specific run or all runs from a player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "map_name",
					Description: "The player run map ('all' for every map)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "nick",
					Description: "The player nickname (case sensitive)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timer",
					Description: "The player run timer (00:00)",
					Required:    false,
				},
			},
		},
		{
			Name:        "zrename",
			Description: "[ADMIN ONLY] Rename a player nickname to a new one",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "old_nick",
					Description: "The old player nickname (case sensitive)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "new_nick",
					Description: "The new player nickname (case sensitive)",
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
	newRunsSevice := automation.NewRunnersService(dg, db, cfg)
	fileUpdaterService := automation.NewFileUpdater(dg, db, cfg)
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
		NewRunners:    newRunsSevice,
		FileUpdater:   fileUpdaterService,
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
	b.NewRunners.Start()
	b.FileUpdater.Start()
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
	case "player_info":
		b.handlePlayerInfoCommand(s, i)
	case "last_runs":
		b.handleLastRunsCommand(s, i)
	case "zadd":
		b.handleAddCommand(s, i)
	case "zremove":
		b.handleRemoveCommand(s, i)
	case "zrename":
		b.handleRenameCommand(s, i)
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
	mapName = helpers.MapNameNormalizer(mapName)
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

func (b *Bot) handlePlayerInfoCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	playerName := optionMap["nick"].StringValue()

	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Printf("[DISCORD] Failed to get channel while handling PlayerInfoCommand: %v", err)
		return
	}

	mapName := channel.Name
	mapName = helpers.MapNameNormalizer(mapName)
	if !helpers.IsValidTable(mapName, b.Config.AllowedMaps) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command can only be used in movement map channels.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send ephemeral message for wrong channel while handling PlayerInfoCommand: %v", err)
		}
		return
	}

	embed := commands.PlayerInfo(b.DB, playerName, mapName, b.Config.AllowedMaps)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to PlayerInfoCommand: %v", err)
		return
	}
}

func (b *Bot) handleLastRunsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	playerName := optionMap["nick"].StringValue()

	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Printf("[DISCORD] Failed to get channel: %v", err)
		return
	}
	mapName := channel.Name
	mapName = helpers.MapNameNormalizer(mapName)
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

	content := commands.LastRuns(b.DB, playerName, mapName, b.Config.AllowedMaps)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to respond to LastRun command: %v", err)
		return
	}

}

func (b *Bot) handleAddCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.IsAdmin(i.Member.User.ID, b.Config.AdminIDs) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nice try, only admins can use this command.",
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send permission denied message: %v", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	playerName := optionMap["nick"].StringValue()
	playerTimer := optionMap["timer"].StringValue()
	mapName := optionMap["map_name"].StringValue()

	content := commands.AddRun(b.DB, playerName, playerTimer, mapName, b.Config.AllowedMaps)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{content},
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to add run: %v", err)
	}
}

func (b *Bot) handleRemoveCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.IsAdmin(i.Member.User.ID, b.Config.AdminIDs) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nice try, only admins can use this command.",
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send permission denied message: %v", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	playerName := optionMap["nick"].StringValue()
	mapName := optionMap["map_name"].StringValue()

	var playerTimer string
	if opt, ok := optionMap["timer"]; ok {
		playerTimer = opt.StringValue()
	}

	content := commands.RemoveRun(b.DB, playerName, playerTimer, mapName, b.Config.AllowedMaps)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{content},
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to remove run(s): %v", err)
	}
}

func (b *Bot) handleRenameCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.IsAdmin(i.Member.User.ID, b.Config.AdminIDs) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Nice try, only admins can use this command.",
			},
		})
		if err != nil {
			log.Printf("[DISCORD] Failed to send permission denied message: %v", err)
		}
		return
	}

	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	oldName := optionMap["old_nick"].StringValue()
	newName := optionMap["new_nick"].StringValue()

	content := commands.RenamePlayer(b.DB, oldName, newName, b.Config.AllowedMaps)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{content},
		},
	})
	if err != nil {
		log.Printf("[DISCORD] Failed to remove run(s): %v", err)
	}
}
