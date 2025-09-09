# GOLDEN SAPLING

A discord bot created to manage the speedrun community of the game Apex Legends.

## Features

### Automations

After setting up the database and game connection, the bot will automatically manage the following tasks:

- **Speedrun Submissions**: Players that complete runs in the game server will have their times automatically submitted to the bot.
- **Players Online Tracking**: The bot keeps track of players currently online in the game server, providing real-time updates to the community.
- **Game server restarts**: The bot monitors the game server and automatically restarts when it goes down.
- **Leaderboard Management**: The bot maintains and updates the leaderboards for the maps in the discord server, ensuring that players can see the latest rankings in real-time.
- **Automatic Bans**: The bot scans messages for common spam/scam words and automatically bans offending users to maintain a safe community environment.
- **Link Fixing**: The bot detects links from platforms like X and Reddit, replying with enhanced versions that provide better media embeds for improved user experience.

### Commands

The bot provides several commands to interact with the speedrun data:

- `/help`: Displays a help message with information about the bot's features and commands.
- `/leaderboard`: Displays the leaderboard for a specified map.
- `/player_info [player]`: Displays information about a specified player in the specific map.
- `/zadd [player] [timer] [map]`: Adds a new run for the specified player on the given map with the provided time.
- `/zremove [player] [map] [timer]`: Remove one or all runs for the specified player on the given map.
- `/zrename [old_player] [new_player]`: Renames a player in the database.

## Setup

### DB

Create a sqlite database with the following information:

- table names should be map names and will be added in the .env for the bot interaction.
- table fields should contain `ID` as an auto-increment primary key, `player_name` as a text field and `time_score` as an integer field.

### BOT

0. Make sure you have Go and a C compiler installed on your machine.
1. Clone the repository.
2. Remove the `.example` from the `.env` file in the root directory and add your environment variables.
3. Run `go build` to compile the bot.
4. Execute the compiled binary to start the bot.
