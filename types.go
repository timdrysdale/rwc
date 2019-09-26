package rwc

import (
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/reconws"
)

type Hub struct {
	Messages  *hub.Hub
	Clients   map[string]*Client //map Id string to client
	Rules     map[string]Rule    //map Id string to Rule
	Add       chan Rule
	Delete    chan string      //Id string
	Broadcast chan hub.Message //for messages incoming from the websocket server(s)
}

type Rule struct {
	Id          string
	Stream      string
	Destination string
}

type Client struct {
	Hub       *Hub //can access messaging hub via <client>.Hub.Messages
	Messages  *hub.Client
	Stopped   chan struct{}
	Websocket *reconws.ReconWs
}
