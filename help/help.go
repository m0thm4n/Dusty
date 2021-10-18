package help

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func HandleHelpCommand(s *discordgo.Session, m *discordgo.Message) {
	message := "```txt\n%s\n%s\n%-16s\t%-20s\n%-16s\t%-20s\n%-20s\t%-20s\n%-16s\t%-20s\n%-16s\t%-20s\n%-16s\t%-20s\n%-16s\t%-20s```"
	message = fmt.Sprintf(message, "Help Information", strings.Repeat("-", len("Help Information")),
		"Provide a song to play from Youtube", "#play",
		"Provide an open.spotify.com URL to play a playlist from Spotify", "#list",
		"Works like play but will return the top 5 results", "#search",
		"Skips a song", "#skip",
		"Stops playing in Discord", "#stop",
		"Shows the current queue", "#show",
		"Shows all playlists for a given user", "#getplaylists <username from spotify>",
	)
	s.ChannelMessageSend(m.ChannelID, message)
}