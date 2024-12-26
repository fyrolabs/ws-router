package ws

import (
	"encoding/json"
	"errors"

	"github.com/olahol/melody"
)

var (
	ResponseSuccess  = H{"success": true}
	EventStatusEnded = "ended"
)

type Route struct {
	// Unique key for the route, used to match incoming requests
	Key     string
	Handler RouteHandler
}

type RouteHandler func(*Context)

type Request struct {
	// Unique ID set by the client, can be empty, used to track responses
	ID string `json:"id"`
	// Action is the name to match a route by a given key string
	Action string `json:"action"`
	// Params is the json data params sent by the client
	Params json.RawMessage `json:"params"`
}

type Reply struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Data   any    `json:"data,omitempty"`
	Error  error  `json:"error,omitempty"`
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
	reply := Reply{
		ID:     c.Request.ID,
		Action: c.Request.Action,
		Data:   msg,
	}

	c.Write(reply)
}

func (c *Context) Error(err error) {
	if !errors.As(err, &Error{}) {
		err = Error{
			Name:    "GenericError",
			Message: err.Error(),
		}
	}

	reply := Reply{Error: err}

	if c.Request != nil {
		reply.ID = c.Request.ID
		reply.Action = c.Request.Action
	}

	c.Write(reply)
}

func (c *Context) Write(data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.session.Write(msg)
}
