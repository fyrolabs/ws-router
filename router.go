package ws

import (
	"net/http"

	"github.com/olahol/melody"
)

type Router struct {
	routes      map[string]RouteHandler
	channels    map[string]*Channel
	BeforeRoute func(c *Context) error
	HandleError func(s *melody.Session, err error)
	mldy        *melody.Melody
}

type H map[string]any

func New(routes []Route, channels []Channel) *Router {
	mldy := melody.New()
	mldy.Config.MaxMessageSize = 500_000 // 500 KB

	routeMap := map[string]RouteHandler{}
	for _, route := range routes {
		routeMap[route.Key] = route.Handler
	}

	channelMap := map[string]*Channel{}
	for _, channel := range channels {
		channelMap[channel.Name] = &channel
	}

	ws := &Router{
		routes:   routeMap,
		channels: channelMap,
		mldy:     mldy,
	}

	ws.setup()
	return ws
}

func (r *Router) SetCheckOrigin(fn func(r *http.Request) bool) {
	r.mldy.Upgrader.CheckOrigin = fn
}

func (r *Router) setup() {
	// Add default watch handler
	r.routes["subscribe"] = r.subscribeRouteHandler()

	r.mldy.HandleMessage(func(s *melody.Session, msg []byte) {
		r.RouteRequest(s, msg)
	})

	r.mldy.HandleError(func(s *melody.Session, err error) {
		if r.HandleError != nil {
			r.HandleError(s, err)
		}
	})

	r.mldy.HandleDisconnect(func(s *melody.Session) {
		// Remove subscribers from channels
		for _, c := range r.channels {
			for _, sessions := range c.Subscribers {
				sessions.Remove(s)
			}
		}
	})
}

func (r *Router) HandleRequest(
	wt http.ResponseWriter, rd *http.Request, data map[string]any,
) {
	r.mldy.HandleRequestWithKeys(wt, rd, data)
}
