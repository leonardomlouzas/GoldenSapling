package commands

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/leonardomlouzas/GoldenSapling/internal/config"
	"github.com/leonardomlouzas/GoldenSapling/internal/helpers"
)

func PlayerInfo(db *sql.DB, playerName, mapName string, allowedMaps []config.MapInfo) *discordgo.MessageEmbed {
	entry := helpers.PlayerInfoReader(db, playerName, mapName, allowedMaps)
	if entry == nil {
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("No records found for this player on %s", mapName),
			Color:       0xffa600,
		}
	}

	return &discordgo.MessageEmbed{
		Title:       mapName,
		Description: fmt.Sprintf("%s statistics:", playerName),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Best Time", Value: fmt.Sprintf("%s (x%s)", entry.BestTime, entry.BestTimeAmount), Inline: true},
			{Name: "Total Runs", Value: fmt.Sprint(entry.TotalRuns), Inline: true},
			{Name: "Worst Time", Value: entry.SlowestRun, Inline: true},
			{Name: "Last Run", Value: entry.LastRun, Inline: true},
			{Name: "First Run", Value: entry.FirstRun, Inline: true},
			{Name: "Total Time", Value: entry.TotalTime, Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "/last_runs to see most recent runs",
		},
	}
}
