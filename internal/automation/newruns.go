package automation

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

type NewRunners struct {
	updateInterval time.Duration
	session        *discordgo.Session
	channelID      string
	folderPath     string
	db             *sql.DB
	allowedMaps    []config.MapInfo
}

func NewRunnersService(s *discordgo.Session, db *sql.DB, cfg *config.Config) *NewRunners {
	if cfg.NewRunsChannelID == "" {
		log.Println("[DISCORD] NEW_RUNNERS_CHANNEL_ID not set, 'New Runners' feature disabled")
		return nil
	}
	return &NewRunners{
		session:        s,
		channelID:      cfg.NewRunsChannelID,
		updateInterval: cfg.UpdateInterval,
		folderPath:     cfg.NewRunsPath,
		db:             db,
		allowedMaps:    cfg.AllowedMaps,
	}
}

func (sc *NewRunners) Start() {
	if sc == nil {
		return // Service is disabled
	}
	log.Println("[DISCORD] Starting 'New Runners'...")

	ticker := time.NewTicker(sc.updateInterval)
	go func() {
		// Perform an initial update on start
		sc.updateNewRunners()
		for range ticker.C {
			sc.updateNewRunners()
		}
	}()
}

func (sc *NewRunners) insertRunsInBatch(mapName string, runs []helpers.NewRunEntry) error {
	if !helpers.IsValidTable(mapName, sc.allowedMaps) {
		log.Printf("[SECURITY] Attempted to insert runs into an invalid table: %s", mapName)
		return nil // Or return an error, but we don't want to stop the whole process
	}

	tx, err := sc.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on any error

	stmt, err := tx.Prepare(`INSERT INTO "` + mapName + `" (player_name, time_score) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, run := range runs {
		timeInt, err := strconv.Atoi(run.TimeScore)
		if err != nil {
			log.Printf("[DISCORD] Invalid time score for %s on map %s, skipping run: %v", run.PlayerName, mapName, err)
			continue // Skip this run, but continue with others in the batch
		}

		if _, err := stmt.Exec(run.PlayerName, timeInt); err != nil {
			// The defer tx.Rollback() will handle this
			return err
		}
	}

	return tx.Commit()
}

func (sc *NewRunners) updateNewRunners() {
	if sc == nil {
		return // Service is disabled
	}

	files, err := os.ReadDir(sc.folderPath)
	if err != nil {
		log.Printf("[DISCORD] Failed to read New Runners directory: %v", err)
		return
	}

	// A map to hold new run entries, keyed by map name.
	runEntries := make(map[string][]helpers.NewRunEntry)
	// A map to hold file paths for each map, to be deleted after successful DB insertion.
	filePathsByMap := make(map[string][]string)

	// 1. Read all files and group runs by map name.
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := sc.folderPath + "/" + file.Name()
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("[DISCORD] Failed to read file %s while in updateNewRunners: %v", file.Name(), err)
			continue
		}
		if len(content) == 0 {
			continue
		}

		entry := helpers.NewRunReader(content)
		runEntries[entry.MapName] = append(runEntries[entry.MapName], entry)
		filePathsByMap[entry.MapName] = append(filePathsByMap[entry.MapName], filePath)
	}

	// 2. Insert runs into the database in batches and send Discord notifications.
	for mapName, newEntries := range runEntries {
		if len(newEntries) == 0 {
			continue
		}

		if err := sc.insertRunsInBatch(mapName, newEntries); err != nil {
			log.Printf("[DISCORD] Failed to insert batch of new runs for map %s: %v. Files will not be deleted.", mapName, err)
			continue // Move to the next map
		}

		// 3. Send notifications to Discord in chunks of 10.
		for len(newEntries) > 10 {
			_, err = sc.session.ChannelMessageSend(sc.channelID, helpers.NewRunTable(newEntries[:10]))
			if err != nil {
				log.Printf("[DISCORD] Failed to send new runs for map %s: %v", mapName, err)
			}
			newEntries = newEntries[10:]
		}
		_, err = sc.session.ChannelMessageSend(sc.channelID, helpers.NewRunTable(newEntries))
		if err != nil {
			log.Printf("[DISCORD] Failed to send new runs for map %s: %v", mapName, err)
		}

		// 4. Delete the processed files.
		for _, path := range filePathsByMap[mapName] {
			if err := os.Remove(path); err != nil {
				log.Printf("[DISCORD] Failed to delete processed run file %s: %v", path, err)
			}
		}
	}
}
