package lavalink

import (
	"bot/config"
	"bot/internal/discord"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

type Lavalink struct {
	Client disgolink.Client
	Queue QueueManager
}

var LavalinkClient *Lavalink


func InitLavalink(){
	LavalinkClient = &Lavalink{Queue: QueueManager{}}

	Lclient := disgolink.New(
		snowflake.MustParse(discord.Session.State.User.ID),
		// TODO: добавить хендлеры для музыки 
	)
	LavalinkClient.Client = Lclient

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	node, err := LavalinkClient.Client.AddNode(ctx, disgolink.NodeConfig{
		Name:     config.GetLavalinkNodeName(),
		Address:  config.GetLavalinkAddr(),
		Password: config.GetLavalinkPass(),
		Secure:   config.GetLavalinkSecure(),
	})

	if err != nil {
		log.Fatal(err)
	}

	version, err := node.Version(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("node version: %s", version)

}