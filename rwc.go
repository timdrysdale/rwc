package rwc

import (
	"time"

	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/reconws"
)

// pass in the messaging hub as a parameter
// assume it is already running
func New(messages *hub.Hub) *Hub {

	h := &Hub{
		Messages: messages,
		Clients:  make(map[string]*Client), //map Id string to Client
		Rules:    make(map[string]Rule),    //map Id string to Rule
		Add:      make(chan Rule),
		Delete:   make(chan string), //Id string
	}

	return h
}

func (h *Hub) Run(closed chan struct{}) {

	defer func() {
		//on exit, go through the list of open clients and close them
		//may panic if a client is individually closed just before exiting
		//but if exiting, a panic is less of an issue
		for _, client := range h.Clients {
			close(client.Websocket.Stop)
		}
	}()

	for {
		select {
		case <-closed:
			return
		case rule := <-h.Add:

			// Allow multiple destinations for a stream;
			// allow multiple streams per destination;
			// allow only one client per rule.Id.
			// Delete any pre-existing client for this rule.Id
			// because it just became superseded
			if client, ok := h.Clients[rule.Id]; ok {
				client = h.Clients[rule.Id]
				close(client.Websocket.Stop) //stop the websocket client
				close(client.Stopped)        //stop RelayIn() & RelayOut()
				delete(h.Clients, rule.Id)
			}
			if _, ok := h.Rules[rule.Id]; ok {
				delete(h.Rules, rule.Id)

			}

			//record the new rule for later convenience in reporting
			h.Rules[rule.Id] = rule

			// create new reconnecting websocket client
			ws := reconws.New()

			ws.Url = rule.Destination //no sanity check - don't dupe ws functionality

			// create client to handle stream messages
			messageClient := &hub.Client{Hub: h.Messages,
				Name:  rule.Destination,
				Topic: rule.Stream,
				Send:  make(chan hub.Message),
				Stats: hub.NewClientStats()}

			client := &Client{Hub: h,
				Messages:  messageClient,
				Stopped:   make(chan struct{}),
				Websocket: ws}

			h.Clients[rule.Id] = client

			h.Messages.Register <- client.Messages //register for messages from hub

			go client.RelayIn()
			go client.RelayOut()
			go ws.Reconnect() //user must check stats to learn of errors
			// an RPC style return on start is of limited value because clients are long lived
			// so we'll need to check the stats later anyway; better just to do things one way

		case ruleId := <-h.Delete:
			if client, ok := h.Clients[ruleId]; ok {
				close(client.Websocket.Stop) //stop the websocket client
				close(client.Stopped)        //stop RelayIn() & RelayOut()
				delete(h.Clients, ruleId)
			}
			if _, ok := h.Rules[ruleId]; ok {
				delete(h.Rules, ruleId)
			}
		}
	}
}

// relay messages from the hub to the websocket client until stopped
func (c *Client) RelayOut() {
	for {
		select {
		case <-c.Stopped:
			break
		case msg, ok := <-c.Messages.Send:
			if ok {
				c.Websocket.Out <- reconws.WsMessage{Data: msg.Data, Type: msg.Type}
			}
		}
	}
}

// relay messages from websocket server to the hub until stopped
func (c *Client) RelayIn() {
	for {
		select {
		case <-c.Stopped:
			break
		case msg, ok := <-c.Websocket.In:
			if ok {
				c.Hub.Messages.Broadcast <- hub.Message{Data: msg.Data, Type: msg.Type, Sender: *c.Messages, Sent: time.Now()}
			}
		}
	}
}
