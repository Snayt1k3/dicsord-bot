package handlers

import (
	"context"
	"log/slog"
	"bot/config"
	"bot/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/snowflake/v2"
)

// ReadyHandler will be called when the bot receives the "ready" event from Discord.
func ReadyHandler(s *discordgo.Session, event *discordgo.Ready) {
	// Set the playing status.
	err := s.UpdateCustomStatus(config.GetBotStatus())
	if err != nil {
		slog.Warn("failed to update game status", "error", err)
	}
}

func OnVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	if event.UserID != session.State.User.ID {
		return
	}

	var channelID *snowflake.ID
	if event.ChannelID != "" {
		id := snowflake.MustParse(event.ChannelID)
		channelID = &id
	}
	discord.Bot.Lavalink.OnVoiceStateUpdate(context.TODO(), snowflake.MustParse(event.GuildID), channelID, event.SessionID)
	
	if event.ChannelID == "" {
		discord.Bot.Queues.Delete(event.GuildID)
	}
}

func OnVoiceServerUpdate(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
	discord.Bot.Lavalink.OnVoiceServerUpdate(context.TODO(), snowflake.MustParse(event.GuildID), event.Token, event.Endpoint)
}

func HelpHandler(session *discordgo.Session, i *discordgo.InteractionCreate) {
	helpMessage := "**🎵 Frieren Bot Help Menu 🎵**\n" +
		"Hello! Here are the commands you can use:\n\n" +
		"**Main Commands:**\n" +
		"- `/play <song_name/link>` – Add a song to the queue and start playing.\n" +
		"- `/pause` – Pause the music.\n" +
		"- `/resume` – Resume playing the music.\n" +
		"- `/stop` – Stop the music and clear the queue.\n" +
		"- `/skip` – Skip the current song.\n\n" +
		
		"**Information:**\n" +
		"- `/help` – Show this help menu.\n\n" +

		"**Notes:**\n" +
		"- Make sure you're in a voice channel before using music commands.\n" +
		"- For questions or suggestions, contact the server administrator.\n\n" +
		"**Thank you for using me!** 🎧"

	discord.Bot.Session.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: helpMessage,
			},
		},
	)
}