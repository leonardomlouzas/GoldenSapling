package automation

import (
	"database/sql"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

type FileUpdater struct {
	updateInterval time.Duration
	session        *discordgo.Session
	folderPath     string
	db             *sql.DB
	allowedMaps    []config.MapInfo
}

func NewFileUpdater(s *discordgo.Session, db *sql.DB, cfg *config.Config) *FileUpdater {
	if cfg.Top10FilePath == "" {
		log.Println("[DISCORD] TOP_10_FILE_PATH not set, 'FileUpdater' feature disabled")
		return nil
	}

	return &FileUpdater{
		updateInterval: cfg.UpdateInterval,
		session:        s,
		folderPath:     cfg.Top10FilePath,
		db:             db,
		allowedMaps:    cfg.AllowedMaps,
	}
}

func (sc *FileUpdater) Start() {
	if sc == nil {
		return // Service is disabled
	}
	log.Println("[DISCORD] Starting 'File Updater'...")

	ticker := time.NewTicker(sc.updateInterval)
	go func() {
		sc.updateFile()
		for range ticker.C {
			sc.updateFile()
		}
	}()
}

func (sc *FileUpdater) updateFile() {
	if sc == nil {
		return // Service is disabled
	}
	helpers.UpdateTop10File(sc.folderPath, sc.db, sc.allowedMaps)
}
