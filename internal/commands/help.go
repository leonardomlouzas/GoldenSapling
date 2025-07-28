package commands

func Help(isDiscordBot bool) string {
	if isDiscordBot {
		return "This is a Discord bot command. Use `/help` to see available commands."
	}
	return "twitch placeholder for help command"
}
