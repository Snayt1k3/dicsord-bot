package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"
	"bot/internal/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

var (
	urlPattern    = regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?")
	searchPattern = regexp.MustCompile(`^(.{2})search:(.+)`)
)


func PlayCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	identifier := i.ApplicationCommandData().Options[0].StringValue()
	
	if !urlPattern.MatchString(identifier) && !searchPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTube.Apply(identifier)
	}

	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You are not in a voice channel!",
			},
		})
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		return err
	}

	player := discord.Bot.Lavalink.Player(snowflake.MustParse(i.GuildID))

	queue := discord.Bot.Queues.Get(i.GuildID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track

	discord.Bot.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loading: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})
			if player.Track() == nil {
				toPlay = &track
			} else {
				queue.Add(track)
			}
		},
		func(playlist lavalink.Playlist) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loading playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})
			if player.Track() == nil {
				toPlay = &playlist.Tracks[0]
				queue.Add(playlist.Tracks[1:]...)
			} else {
				queue.Add(playlist.Tracks...)
			}
		},
		func(tracks []lavalink.Track) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loading: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
			})
			if player.Track() == nil {
				toPlay = &tracks[0]
			} else {
				queue.Add(tracks[0])
			}
		},
		func() {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("No matches: `%s`", identifier)),
			})
		},
		func(err error) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Error while searching: `%s`", err)),
			})
		},
	))
	if toPlay == nil {
		return nil
	}

	if err := s.ChannelVoiceJoinManual(i.GuildID, voiceState.ChannelID, false, false); err != nil {
		return err
	}
	
	return player.Update(context.TODO(), lavalink.WithTrack(*toPlay))

}

func StopCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	queue := discord.Bot.Queues.Get(i.GuildID)
	queue.Clear()

	voiceState, _ := s.State.VoiceState(i.GuildID, s.State.User.ID)
	if voiceState == nil || voiceState.ChannelID == "" {
		return
	}

	err := s.ChannelVoiceJoinManual(i.GuildID, "", false, false)

	if err != nil {
		slog.Error("Error while disconnecting: ", "error", err)
	}

}

func SkipCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	queue := discord.Bot.Queues.Get(i.GuildID)
	track, exists := queue.Next()

	if !exists {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "The queue is empty. Add a new song using the /play command.",
			},
		})
	}

	player := discord.Bot.Lavalink.Player(snowflake.MustParse(i.GuildID))

	if err := player.Update(context.TODO(), lavalink.WithTrack(track)); err != nil {
		slog.Error("Failed to play next track: ", "error", err)
	}

	s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
		Title:       "Now is playing 🎶",
		Description: fmt.Sprintf("[%s](%s)", track.Info.Title, *track.Info.URI),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: *track.Info.ArtworkURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  track.Info.Length.String(),
				Inline: true,
			},
		},
	})

	return nil

}

func PauseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	player := discord.Bot.Lavalink.Player(snowflake.MustParse(i.GuildID))

	if player == nil {
		return nil
	}

	if player.Paused() {
		return nil
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(false)); err != nil {
		return discord.Bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error while pausing",
			}})
	}

	return discord.Bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Player is now Paused",
		},
	})
}
func ResumeHandler(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	player := discord.Bot.Lavalink.Player(snowflake.MustParse(i.GuildID))

	if player == nil {
		return nil
	}

	if !player.Paused() {
		return nil
	}

	if err := player.Update(context.TODO(), lavalink.WithPaused(true)); err != nil {
		return discord.Bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error while resuming",
			}})
	}

	return discord.Bot.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Player is now Resumed",
		},
	})
}
