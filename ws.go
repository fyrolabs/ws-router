package ws

import (
	"net/http"

	"github.com/olahol/melody"
)

type WS struct {
	Routes      Routes
	Channels    Channels
	BeforeRoute func(c *Context) error
	HandleError func(s *melody.Session, err error)
	mldy        *melody.Melody
}

type Channels map[string]*Channel
type Routes map[string]RouteHandler

type H map[string]any

func New(routes Routes, channels Channels) *WS {
	mldy := melody.New()
	mldy.Config.MaxMessageSize = 500_000 // 500 KB

	ws := &WS{
		Routes:   routes,
		Channels: channels,
		mldy:     mldy,
	}

	ws.setup()
	return ws
}

func (w *WS) SetCheckOrigin(fn func(r *http.Request) bool) {
	w.mldy.Upgrader.CheckOrigin = fn
}

func (w *WS) setup() {
	// Add default watch handler
	w.Routes["subscribe"] = w.SubscribeHandler()

	w.mldy.HandleMessage(func(s *melody.Session, msg []byte) {
		w.RouteRequest(s, msg)
	})

	w.mldy.HandleError(func(s *melody.Session, err error) {
		if w.HandleError != nil {
			w.HandleError(s, err)
		}
	})

	w.mldy.HandleDisconnect(func(s *melody.Session) {
		// Remove subscribers from channels
		for _, c := range w.Channels {
			for _, sessions := range c.Subscribers {
				sessions.Remove(s)
			}
		}
	})
}

func (w *WS) HandleRequest(
	wt http.ResponseWriter, rd *http.Request, keys map[string]any,
) {
	w.mldy.HandleRequestWithKeys(wt, rd, keys)
}
