package ws

import (
	"encoding/json"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/olahol/melody"
)

type Channel struct {
	Name        string
	KeyFunc     ChannelKeyFunc
	Subscribers Subscribers
}

type ChannelKeyFunc func(*Context, json.RawMessage) (string, error)
type Subscribers map[string]mapset.Set[*melody.Session]

// NewChannel creates a new channel with the given name and optional key function
// which returns a string.
// The key function is used to determine the key to seperate subscribers within a channel.
// This can be an user ID, or an enum string to separate different actions within a channel.
// If the keyFunc is nil, then no separation will occur.
func NewChannel(name string, keyFunc ChannelKeyFunc) *Channel {
	return &Channel{
		Name:        name,
		KeyFunc:     keyFunc,
		Subscribers: Subscribers{},
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
			}
		}

		channel.addSubscriber(key, c.session)
		c.Respond(ResponseSuccess)
	}
}

func (c *Channel) addSubscriber(key string, s *melody.Session) {
	_, exists := c.Subscribers[key]
	if !exists {
		c.Subscribers[key] = mapset.NewSet[*melody.Session]()
	}

	c.Subscribers[key].Add(s)
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

	sessions, exists := c.Subscribers[key]
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
