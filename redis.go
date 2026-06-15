package ws

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisBroadcaster struct {
	router   *Router
	rdb      *redis.Client
	rChannel string
}

type RedisBroadcasterOpts struct {
	Router       *Router
	RDB          *redis.Client
	RedisChannel string
}

func NewRedisBroadcaster(opts RedisBroadcasterOpts) *RedisBroadcaster {
	return &RedisBroadcaster{
		router:   opts.Router,
		rdb:      opts.RDB,
		rChannel: opts.RedisChannel,
	}
}

func (b *RedisBroadcaster) Server() {
	pubSub := b.rdb.Subscribe(context.Background(), b.rChannel)
	redisChan := pubSub.Channel()

	for {
		redisMsg := <-redisChan
		var opts BroadcastOpts
		if err := json.Unmarshal([]byte(redisMsg.Payload), &opts); err != nil {
			log.Println(err)
			continue
		}

		if err := b.router.Broadcast(opts); err != nil {
			log.Println(err)
		}
	}
}

func (b *RedisBroadcaster) Broadcast(opts BroadcastOpts) error {
	payload, err := json.Marshal(opts)
	if err != nil {
		return err
	}

	return b.rdb.Publish(context.Background(), b.rChannel, payload).Err()
}
