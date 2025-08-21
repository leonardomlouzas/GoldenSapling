package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func PlayerInfo(db *sql.DB, playerName string, mapName string, allowedMaps []config.MapInfo) *discordgo.MessageEmbed {
	entry := helpers.PlayerInfoReader(db, playerName, mapName, allowedMaps)
	if entry == nil {
		return &discordgo.MessageEmbed{
			Description: "No records found for this player on the specified map.",
			Color:       0xffa600,
		}
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Information", playerName),
		Description: fmt.Sprintf("Overall player info on map %s", mapName),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Best Time", Value: entry.BestTime, Inline: true},
			{Name: "Total Runs", Value: fmt.Sprint(entry.TotalRuns), Inline: true},
			{Name: "Last Run", Value: entry.LastRun, Inline: true},
			{Name: "First Run", Value: entry.FirstRun, Inline: true},
			{Name: "Total Time", Value: entry.TotalTime, Inline: true},
			{Name: "Best Time Amount", Value: entry.BestTimeAmount, Inline: true},
		},
	}
}
