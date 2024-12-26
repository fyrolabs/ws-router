package ws

import (
	"encoding/json"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/olahol/melody"
)

type Channel struct {
	Name        string
	KeyFunc     ChannelKeyFunc
	subscribers subscribers
}

type ChannelKeyFunc func(ctx *Context, params json.RawMessage) (string, error)
type subscribers map[string]mapset.Set[*melody.Session]

// NewChannel creates a new channel with the given name and optional key function
// which returns a string.
// The key function is used to determine the key to seperate subscribers within a channel.
// This can be an user ID, or an enum string to separate different actions within a channel.
// If the keyFunc is nil, then no separation will occur.
func NewChannel(name string, keyFunc ChannelKeyFunc) *Channel {
	return &Channel{
		Name:        name,
		KeyFunc:     keyFunc,
		subscribers: subscribers{},
	}
}

type SubscribeRequest struct {
	Channel string          `json:"channel"`
	Params  json.RawMessage `json:"params"`
}

type BroadcastReply struct {
	Channel string `json:"channel"`
	Event   string `json:"event,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   error  `json:"error,omitempty"`
}

// subscribeRouteHandler processes a subscribe request sent by the client.
// keyFunc is run to obtain a sub key to separate subscribers within the channel.
// if an error occurs during keyFunc(), it is returned to the client and the client is not subscribed.
func (r *Router) subscribeRouteHandler() RouteHandler {
	return func(c *Context) {
		var req SubscribeRequest
		if err := json.Unmarshal(c.Request.Params, &req); err != nil {
			c.Error(err)
			return
		}

		channel, exists := r.channels[req.Channel]
		if !exists {
			c.Error(ErrChannelNotFound)
			return
		}

		var key string
		if channel.KeyFunc == nil {
			key = ""
		} else {
			var err error
			key, err = channel.KeyFunc(c, req.Params)
			if err != nil {
				c.Error(err)
				return
			}
		}

		channel.addSubscriber(key, c.session)
		c.Respond(ResponseSuccess)
	}
}

func (c *Channel) addSubscriber(key string, s *melody.Session) {
	_, exists := c.subscribers[key]
	if !exists {
		c.subscribers[key] = mapset.NewSet[*melody.Session]()
	}

	c.subscribers[key].Add(s)
}

// Broadcast sends a message to all subscribers in the channel with the given key.
// If the channel is not key separated, then the key can be an empty string.
// event: name of the event to send to the client
// msg: JSON serializable payload
func (r *Router) Broadcast(
	channel string, key string, event string, msg any,
) error {
	c, exists := r.channels[channel]
	if !exists {
		return ErrChannelNotFound
	}

	reply := BroadcastReply{Channel: channel, Event: event, Data: msg}
	payload, err := json.Marshal(reply)
	if err != nil {
		return err
	}

	sessions, exists := c.subscribers[key]
	if !exists {
		return nil // No subscribers, return early
	}

	it := sessions.Iterator()
	for session := range it.C {
		if err := session.Write(payload); err != nil {
			return err
		}
	}

	return nil
}
