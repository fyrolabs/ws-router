package ws

import (
	"encoding/json"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/olahol/melody"
)

var (
	ErrChannelNotFound = Error{
		Code:    "channel_not_found",
		Message: "channel not found",
	}
)

type Channel struct {
	Name        string
	Handler     ChannelHandler
	Subscribers Subscribers
}

type ChannelHandler func(*Context, json.RawMessage) (string, error)
type Subscribers map[string]mapset.Set[*melody.Session]

func NewChannel(name string, handler ChannelHandler) *Channel {
	return &Channel{
		Name:        name,
		Handler:     handler,
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

func (w *WS) SubscribeHandler() RouteHandler {
	handler := func(c *Context) {
		var req SubscribeRequest
		if err := json.Unmarshal(c.Request.Params, &req); err != nil {
			c.Error(err)
			return
		}

		channel, exists := w.Channels[req.Channel]
		if !exists {
			c.Error(ErrChannelNotFound)
			return
		}

		var key string
		if channel.Handler == nil {
			key = ""
		} else {
			var err error
			key, err = channel.Handler(c, req.Params)
			if err != nil {
				c.Error(err)
			}
		}

		channel.AddSubscriber(key, c.session)
		c.Respond(ResponseSuccess)
	}
	return handler
}

func (c *Channel) AddSubscriber(key string, s *melody.Session) {
	_, exists := c.Subscribers[key]
	if !exists {
		c.Subscribers[key] = mapset.NewSet[*melody.Session]()
	}

	c.Subscribers[key].Add(s)
}

func (w *WS) Broadcast(
	channel string, key string, event string, msg any,
) error {
	c, exists := w.Channels[channel]
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
