package commands

import "github.com/bwmarrin/discordgo"

/*
Returns information about the bot in a Discord embed format.
*/
func HelpDiscord() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "GoldenSapling General Purpose Bot",
		Description: "The main functionality of this bot is to manage the speedruns made in the Movement HUB, but it's not limited to that. It also has other functionalities like automated moderation and better embed links.",
		Color:       0xffa600,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "/help", Value: "Shows this help message."},
			{Name: "Slash commands", Value: "Type a '/' in the chat and select the Golden Sapling bot to see all available commands."},
			{Name: "Link Fixing", Value: "This bot looks for X (Twitter) and Reddit links and replies with a version that provides a better media embed."},
			{Name: "Automatic bans", Value: "This bot looks for common spam/scam words sent by users and automatically bans them."},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "Made by: L"},
	}
}
