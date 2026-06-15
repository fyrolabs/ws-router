package ws

import (
	"encoding/json"
	"errors"

	"github.com/olahol/melody"
)

type SendKind string

const (
	SendKindResponse  SendKind = "response"
	SendKindBroadcast SendKind = "broadcast"
)

var (
	ResponseSuccess = H{"success": true}
)

type Route struct {
	// Unique key for the route, used to match incoming requests
	Key     string
	Handler RouteHandler
}

type RouteHandler func(*Context)

// Request is the incoming message from the client
type Request struct {
	// Unique ID set by the client, can be empty, used to track responses
	ID string `json:"id"`
	// Action is the name to match a route by a given key string
	Action string `json:"action"`
	// Params is the json data params sent by the client
	Params json.RawMessage `json:"params"`
}

type SendData struct {
	Kind SendKind `json:"kind"`
	Data any      `json:"data"`
}

type SendResponse struct {
	RequestID string `json:"requestId,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     error  `json:"error,omitempty"`
}

type SendBroadcast struct {
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Data    any    `json:"data"`
}

func (r *Router) RouteRequest(s *melody.Session, msg []byte) {
	c := &Context{session: s}

	var req *Request
	if err := json.Unmarshal(msg, &req); err != nil {
		c.Error(err)
		return
	}
	c.Request = req

	handler, exists := r.routes[req.Action]
	if !exists {
		c.Error(NewActionNotFoundErr(req.Action))
		return
	}

	// Run before route hook
	if r.BeforeRoute != nil {
		if err := r.BeforeRoute(c); err != nil {
			c.Error(err)
		}
	}

	handler(c)
}

func (c *Context) Respond(msg any) {
	sendData := SendData{
		Kind: SendKindResponse,
		Data: SendResponse{
			RequestID: c.Request.ID,
			Data:      msg,
		},
	}

	c.Write(sendData)
}

func (c *Context) Error(err error) {
	if c.Request == nil {
		return
	}

	if !errors.As(err, &Error{}) {
		err = Error{
			Name:    "GenericError",
			Message: err.Error(),
		}
	}

	sendData := SendData{
		Kind: SendKindResponse,
		Data: SendResponse{
			RequestID: c.Request.ID,
			Error:     err,
		},
	}

	c.Write(sendData)
}

func (c *Context) Write(data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.session.Write(msg)
}
