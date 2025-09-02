package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func RemoveRun(db *sql.DB, playerName string, timer, mapName string, allowedMaps []config.MapInfo) *discordgo.MessageEmbed {
	if !helpers.IsValidTable(mapName, allowedMaps) && mapName != "all" {
		return &discordgo.MessageEmbed{
			Title:       "FAILED",
			Description: "Invalid map.",
			Color:       0xff0000,
		}
	}
	if len(playerName) <= 0 {
		return &discordgo.MessageEmbed{
			Title:       "FAILED",
			Description: "Invalid player name.",
			Color:       0xff0000,
		}
	}
	timerSec := helpers.ConvetTimerToSeconds(timer)

	if timerSec != 0 {
		err := helpers.RemoveRunFromTable(db, mapName, playerName, timerSec, allowedMaps)

		if err != nil {
			return &discordgo.MessageEmbed{
				Title:       "FAILED",
				Description: "DB Removal failed",
				Color:       0xff0000,
			}
		}

		return &discordgo.MessageEmbed{
			Title:       "SUCCESS",
			Description: fmt.Sprintf("%s - %d", playerName, timerSec),
			Color:       0x00ff00,
			Footer: &discordgo.MessageEmbedFooter{
				Text: mapName,
			},
		}
	} else {

		err := helpers.DeleteRunsFromTables(db, mapName, playerName, allowedMaps)

		if err != nil {
			return &discordgo.MessageEmbed{

				Title:       "FAILED",
				Description: "DB Removal failed",
				Color:       0xff0000,
			}
		}

		return &discordgo.MessageEmbed{
			Title:       "SUCCESS",
			Description: fmt.Sprintf("%s", playerName),
			Color:       0x00ff00,
			Footer: &discordgo.MessageEmbedFooter{
				Text: mapName,
			},
		}
	}
}
