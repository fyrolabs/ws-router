# WSRouter

## Routes
```go
type Route struct {
	// Unique key for the route, used to match incoming requests
	Key     string
	Handler RouteHandler
}
```

Define routes with a list of string keys and route handlers.

Note the "subscribe" key is used by channel subscription by default.

## Channels

Create channels by:

```go
// NewChannel creates a new channel with the given name and optional key function
// which returns a string.
// The key function is used to determine the key to seperate subscribers within a channel.
// This can be an user ID, or an enum string to separate different actions within a channel.
// If the keyFunc is nil, then no separation will occur.
func NewChannel(name string, keyFunc ChannelKeyFunc)

// Where keyFunc is:
func(ctx *Context, params json.RawMessage) (string, error)

// ctx is the session context, and params is sent by the subscription request
```

Pass into init

## Init
Create a new instance by:

`ws.New(routes []Route, channels []Channel)`

## Usage with gin

Handle a request within a gin route

```go
wsr := ws.New(...)
// ...

func WebsocketHandler(c *gin.Context) {
  // ... perform auth, get user and associated data

  user := User{
    ID: 1
  }

  data :=  map[string]any{
    "userID": 1,
    "locale": "en"
  }

  wsr.HandleRequest(c.Writer, c.Request, data)
}
```

## Route handling

Client side sends a request
```go
type Request struct {
	// Unique ID set by the client, can be empty, used to track responses
	ID string `json:"id"`
	// Action is the name to match a route by a given key string
	Action string `json:"action"`
	// Params is the json data params sent by the client
	Params json.RawMessage `json:"params"`
}
```

Client sends to a specific route key, which is matched by the action name
Handler code:

```go
// Context is the context for a request.
// use c.Get(key) and c.Set(key) to get and set session data
// c.Respond(data) to send a response
type Context struct {
	Request *Request
}
```

```go
func(c *Context) {
  c.Get(key) // Gets session data from the data passed in from gin websocket handler
  c.Set(key, value) // sets session data

  c.Respond(...) // Response sends back the client specified ID
  c.Error(...) // Send an error to the client
}
```

## Broadcast (Channels)
Broadcast to all subscribers in a channnel:

```go
func (w *WS) Broadcast(
	channel string, key string, event string, msg any,
) error
```

The sent payload is as follows:
```go
type BroadcastReply struct {
	Channel string `json:"channel"`
	Event   string `json:"event,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   error  `json:"error,omitempty"`
}
```
