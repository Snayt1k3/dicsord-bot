package discord

import (
	"context"
	"log/slog"
	"time"

	"bot/config"
	

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

type DiscordBot struct {
	Session *discordgo.Session
	Lavalink disgolink.Client
	Queues   *QueueManager
}

var Bot *DiscordBot

func InitBot() {
	Bot = &DiscordBot{Queues: &QueueManager{Queues: make(map[string]*Queue)}}

	session, err := discordgo.New("Bot " + config.GetDiscordToken()) // Initializing discord session
	session.State.TrackVoice = true
	session.Identify.Intents = discordgo.IntentsAll
	Bot.Session = session

	if err != nil {
		slog.Error("failed to create discord session", "error", err)
	}

}

func InitLavalink() {
	Bot.Lavalink = disgolink.New(snowflake.MustParse(Bot.Session.State.User.ID),
		// disgolink.WithListenerFunc(b.onPlayerPause),
		// disgolink.WithListenerFunc(b.onPlayerResume),
		// disgolink.WithListenerFunc(b.onTrackStart),
		// disgolink.WithListenerFunc(b.onTrackEnd),
		// disgolink.WithListenerFunc(b.onTrackException),
		// disgolink.WithListenerFunc(b.onTrackStuck),
		// disgolink.WithListenerFunc(b.onWebSocketClosed),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Bot.Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     config.GetLavalinkNodeName(),
		Address:  config.GetLavalinkAddr(),
		Password: config.GetLavalinkPass(),
		Secure:   config.GetLavalinkSecure(),
	})

}

func InitConnection() {
	if err := Bot.Session.Open(); err != nil {
		slog.Error("failed to create websocket connection to discord", "error", err)
		return
	}
}