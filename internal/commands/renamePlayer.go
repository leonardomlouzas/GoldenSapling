package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func RenamePlayer(db *sql.DB, oldName, newName string, allowedMaps []config.MapInfo) *discordgo.MessageEmbed {

	err := helpers.RenameRunsFromTables(db, oldName, newName, allowedMaps)

	if err != nil {
		return &discordgo.MessageEmbed{

			Title:       "FAILED",
			Description: "Player rename failed",
			Color:       0xff0000,
		}
	}

	return &discordgo.MessageEmbed{
		Title:       "SUCCESS",
		Description: fmt.Sprintf("%s -> %s", oldName, newName),
		Color:       0x00ff00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "all",
		},
	}

}
