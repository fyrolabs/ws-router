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

type RouteHandler func(*Context)

type Request struct {
	ID     string          `json:"id"`
	Action string          `json:"action"`
	Params json.RawMessage `json:"params"`
}

type Reply struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Data   any    `json:"data,omitempty"`
	Error  error  `json:"error,omitempty"`
}

type Context struct {
	Request *Request
	session *melody.Session
}

func (c *Context) Set(key string, value any) {
	c.session.Set(key, value)
}

func (c *Context) Get(key string) any {
	value, exists := c.session.Get(key)
	if !exists {
		return nil
	}
	return value
}

func (w *WS) RouteRequest(s *melody.Session, msg []byte) {
	c := &Context{session: s}

	var req *Request
	if err := json.Unmarshal(msg, &req); err != nil {
		c.Error(err)
		return
	}
	c.Request = req

	handler, exists := w.Routes[req.Action]
	if !exists {
		c.Error(NewActionNotFoundErr(req.Action))
		return
	}

	// Run before route hook
	if w.BeforeRoute != nil {
		if err := w.BeforeRoute(c); err != nil {
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
