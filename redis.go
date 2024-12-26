package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisBroadcast struct {
	router   *Router
	rdb      *redis.Client
	rChannel string
}

type RedisBroadcastOpts struct {
	Router       *Router
	RDB          *redis.Client
	RedisChannel string
}

type RedisBroadcastData struct {
	Channel string `json:"channel"`
	Key     string `json:"key"`
	Event   string `json:"event"`
	Message any    `json:"message"`
}

func NewRedisBroadcast(opts RedisBroadcastOpts) *RedisBroadcast {
	return &RedisBroadcast{
		router:   opts.Router,
		rdb:      opts.RDB,
		rChannel: opts.RedisChannel,
	}
}

func (b *RedisBroadcast) Server() {
	pubSub := b.rdb.Subscribe(context.Background(), b.rChannel)
	redisChan := pubSub.Channel()

	for {
		redisMsg := <-redisChan
		var data RedisBroadcastData
		if err := json.Unmarshal([]byte(redisMsg.Payload), &data); err != nil {
			log.Println(err)
			continue
		}

		if err := b.router.Broadcast(
			data.Channel, data.Key, data.Event, data.Message,
		); err != nil {
			log.Println(err)
		}
	}
}

func (b *RedisBroadcast) Broadcast(params RedisBroadcastData) error {
	payload, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return b.rdb.Publish(context.Background(), b.rChannel, payload).Err()
}
