# WSRouter

## Routes
`Routes map[string]RouteHandler`

Define routes with a list of string keys and route handlers.

Note the "subscribe" route is used by channel subscription by default.

## Channels

Create channels by:

`ws.NewChannel(name string, handler ws.ChannelHandler)`

Pass into init

## Init
Create a new instance by:

`ws.New(routes Routes, channels Channels)`
