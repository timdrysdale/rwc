package rwc

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/reconws"
)

func init() {

	log.SetLevel(log.InfoLevel)

}

func TestInstantiateHub(t *testing.T) {

	mh := agg.New()

	h := New(mh)

	if reflect.TypeOf(h.Broadcast) != reflect.TypeOf(make(chan hub.Message)) {
		t.Error("Hub.Broadcast channel of wrong type")
	}
	if reflect.TypeOf(h.Clients) != reflect.TypeOf(make(map[string]*Client)) {
		t.Error("Hub.Clients map of wrong type")
	}
	if reflect.TypeOf(h.Add) != reflect.TypeOf(make(chan Rule)) {
		t.Error("Hub.Add channel of wrong type")
	}

	if reflect.TypeOf(h.Delete) != reflect.TypeOf(make(chan string)) {
		t.Errorf("Hub.Delete channel of wrong type wanted/got %v %v", reflect.TypeOf(""), reflect.TypeOf(h.Delete))
	}

	if reflect.TypeOf(h.Rules) != reflect.TypeOf(make(map[string]Rule)) {
		t.Error("Hub.Broadcast channel of wrong type")
	}

}

func TestAddRule(t *testing.T) {

	mh := agg.New()
	h := New(mh)

	closed := make(chan struct{})
	defer close(closed)

	go h.Run(closed)

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s.Close()

	id := "rule0"
	stream := "/stream/large"
	destination := "ws" + strings.TrimPrefix(s.URL, "http") //s.URL //"ws://localhost:8081"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id].Destination != destination {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination, h.Rules[id].Destination)
		}
		if h.Rules[id].Stream != stream {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream, h.Rules[id].Stream)
		}
	}
}

func TestAddRules(t *testing.T) {

	suppressLog()
	defer displayLog()

	closed := make(chan struct{})
	defer close(closed)

	mh := agg.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	// Create test server with the echo handler.
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s1.Close()

	// Create test server with the echo handler.
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s2.Close()

	id := "rule0"
	stream := "/stream/large"
	destination := "ws" + strings.TrimPrefix(s1.URL, "http") //s1.URL //"ws://localhost:8081"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id2 := "rule1"
	stream2 := "/stream/medium"
	destination2 := "ws" + strings.TrimPrefix(s2.URL, "http") //s2.URL //"ws://localhost:8082"

	r2 := &Rule{Id: id2,
		Stream:      stream2,
		Destination: destination2}

	h.Add <- *r2

	time.Sleep(time.Millisecond)

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id].Destination != destination {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination, h.Rules[id].Destination)
			fmt.Printf("%v\n", h.Rules)
		}
		if h.Rules[id].Stream != stream {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream, h.Rules[id].Stream)
			fmt.Printf("%v\n", h.Rules)
		}
	}
	if _, ok := h.Rules[id2]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id2].Destination != destination2 {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination2, h.Rules[id2].Destination)
			fmt.Printf("%v\n", h.Rules)
		}
		if h.Rules[id2].Stream != stream2 {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream2, h.Rules[id2].Stream)
		}
	}

	time.Sleep(500 * time.Millisecond)

}

func TestAddDupeRule(t *testing.T) {

	suppressLog()
	defer displayLog()

	closed := make(chan struct{})
	defer close(closed)

	mh := agg.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	// Create test server with the echo handler.
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s1.Close()

	// Create test server with the echo handler.
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s2.Close()

	id := "rule0"
	stream := "/stream/large"
	destination := "ws" + strings.TrimPrefix(s1.URL, "http") //s.URL //"ws://localhost:8082"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id = "rule0"
	stream = "/stream/medium"
	destination = "ws" + strings.TrimPrefix(s2.URL, "http") //://localhost:8082"

	r = &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id].Destination != destination {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination, h.Rules[id].Destination)
		}
		if h.Rules[id].Stream != stream {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream, h.Rules[id].Stream)
		}
	}

}

func TestDeleteRule(t *testing.T) {
	suppressLog()
	defer displayLog()

	closed := make(chan struct{})
	defer close(closed)

	mh := agg.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	// Create test server with the echo handler.
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s1.Close()

	// Create test server with the echo handler.
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s2.Close()

	id := "rule0"
	stream := "/stream/large"
	destination := "ws" + strings.TrimPrefix(s1.URL, "http") //s1.URL //"ws://localhost:8081"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id2 := "rule1"
	stream2 := "/stream/medium"
	destination2 := "ws" + strings.TrimPrefix(s2.URL, "http") //s2.URL //"ws://localhost:8082"

	r2 := &Rule{Id: id2,
		Stream:      stream2,
		Destination: destination2}

	h.Add <- *r2

	time.Sleep(time.Millisecond)

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id].Destination != destination {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination, h.Rules[id].Destination)
			fmt.Printf("%v\n", h.Rules)
		}
		if h.Rules[id].Stream != stream {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream, h.Rules[id].Stream)
			fmt.Printf("%v\n", h.Rules)
		}
	}
	if _, ok := h.Rules[id2]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id2].Destination != destination2 {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination2, h.Rules[id2].Destination)
			fmt.Printf("%v\n", h.Rules)
		}
		if h.Rules[id2].Stream != stream2 {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream2, h.Rules[id2].Stream)
		}
	}

	h.Delete <- id

	time.Sleep(time.Millisecond)

	if _, ok := h.Rules[id]; ok {
		t.Error("Deleted rule registered in Rules")

	}

	if _, ok := h.Rules[id2]; !ok {
		t.Error("Rule not registered in Rules")

	} else {

		if h.Rules[id2].Destination != destination2 {
			t.Errorf("Rule has incorrect destination wanted/got %v %v\n", destination2, h.Rules[id2].Destination)
			fmt.Printf("%v\n", h.Rules)
		}
		if h.Rules[id2].Stream != stream2 {
			t.Errorf("Rule has incorrect stream wanted/got %v %v\n", stream2, h.Rules[id2].Stream)
		}
	}

}

func TestSendMessageTopic(t *testing.T) {

	time.Sleep(time.Second)
	fmt.Println("-----------------------------------------------")

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shout(w, r)
	}))
	defer s.Close()

	closed := make(chan struct{})
	defer close(closed)

	mh := agg.New()
	go mh.Run(closed)

	time.Sleep(time.Millisecond)

	h := New(mh)
	go h.Run(closed)

	fmt.Printf("Before config : Streams: %v\n", mh.Streams)

	id := "rule0"
	stream := "medium"
	destination := "ws" + strings.TrimPrefix(s.URL, "http")

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(10 * time.Millisecond)

	fmt.Printf("after add rule : Streams: %v\n", mh.Streams)
	fmt.Printf("after add rule : Rules: %+v\n", mh.Rules)

	time.Sleep(time.Millisecond)

	fmt.Println("added rule")

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")

	}

	reply := make(chan hub.Message)

	c := &hub.Client{Hub: mh.Hub, Name: "a", Topic: stream, Send: reply}

	h.Messages.Register <- c

	fmt.Println("registered client")

	time.Sleep(5 * time.Millisecond)

	fmt.Printf("Streams: %v\n", h.Messages.Streams)
	fmt.Printf("SubClients: %v\n", h.Messages.SubClients)
	fmt.Printf("Clients: %v\n", h.Messages.Hub.Clients)

	//if len(mh.Hub.Clients) != 1 {
	//	t.Errorf("Wrong number of clients registered %d", len(mh.SubClients))
	//} else {
	//	fmt.Printf("Clients are registered in quanity of %d\n", len(mh.SubClients))
	//}

	payload := []byte("test message")
	shoutedPayload := []byte("TEST MESSAGE")

	mh.Broadcast <- hub.Message{Data: payload, Type: websocket.TextMessage, Sender: *c, Sent: time.Now()}

	fmt.Println("broadcast messsage")

	//if len(mh.Hub.Clients) != 1 {
	//	t.Errorf("Wrong number of clients registered %d", len(mh.
	//		SubClients))
	//}
	//
	select {
	case msg := <-reply:
		fmt.Println("got reply")
		if bytes.Compare(msg.Data, shoutedPayload) != 0 {
			t.Error("Got wrong message")
		}

	case <-time.After(10 * time.Millisecond):
		t.Error("timed out waiting for message")
	}

	time.Sleep(time.Second)

}

func TestSendMessageStream(t *testing.T) {

	time.Sleep(time.Second)
	fmt.Println("-----------------------------------------------")

	wsMsg := make(chan reconws.WsMessage)

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report(w, r, wsMsg)
	}))
	defer s.Close()

	closed := make(chan struct{})
	defer close(closed)

	mh := agg.New()
	go mh.Run(closed)

	time.Sleep(time.Millisecond)

	h := New(mh)
	go h.Run(closed)

	fmt.Printf("Before config : Streams: %v\n", mh.Streams)

	id := "rule0"
	stream := "/stream/medium"
	destination := "ws" + strings.TrimPrefix(s.URL, "http")

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(10 * time.Millisecond)

	fmt.Printf("after add rule : Streams: %v\n", mh.Streams)
	fmt.Printf("after add rule : Rules: %+v\n", mh.Rules)

	time.Sleep(time.Millisecond)

	fmt.Println("added rule")

	if _, ok := h.Rules[id]; !ok {
		t.Error("Rule not registered in Rules")
	}

	// add rule for stream
	feeds := []string{"video0", "audio"}
	streamRule := &agg.Rule{Stream: stream, Feeds: feeds}

	mh.Add <- *streamRule

	// add clients/feeds to supply the stream
	reply0 := make(chan hub.Message)
	reply1 := make(chan hub.Message)

	c0 := &hub.Client{Hub: mh.Hub, Name: "video0", Topic: feeds[0], Send: reply0}
	c1 := &hub.Client{Hub: mh.Hub, Name: "audio", Topic: feeds[1], Send: reply1}

	h.Messages.Register <- c0
	h.Messages.Register <- c1

	fmt.Println("registered clients")

	time.Sleep(1 * time.Millisecond)

	//fmt.Printf("Streams: %v\n", h.Messages.Streams)
	//fmt.Printf("SubClients: %v\n", h.Messages.SubClients)
	//fmt.Printf("Clients: %v\n", h.Messages.Hub.Clients)

	payload := []byte("test message")
	shoutedPayload := []byte("TEST MESSAGE")

	// client sends a message ...
	c0.Hub.Broadcast <- hub.Message{Data: payload, Type: websocket.TextMessage, Sender: *c0, Sent: time.Now()}
	c1.Hub.Broadcast <- hub.Message{Data: shoutedPayload, Type: websocket.TextMessage, Sender: *c1, Sent: time.Now()}
	time.Sleep(time.Millisecond)

	fmt.Println("broadcast messsages")

	// streams do not send replies to feeds!

	gotMsg := []bool{false, false}
	msgs := [][]byte{payload, shoutedPayload}

	for i := 0; i < 2; i++ {
		select {
		case msg := <-wsMsg:
			fmt.Printf("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&:%s\n", msg.Data)
			for j, contents := range msgs {
				if bytes.Compare(msg.Data, contents) == 0 {
					gotMsg[j] = true
				}
			}
		case <-time.After(5 * time.Millisecond):
			t.Errorf("Timeout on getting %dth websocket message on", i)
		}

	}

	for j, result := range gotMsg {
		if result == false {
			t.Errorf("Did not get %dth websocket mssage", j)
		}
	}

	time.Sleep(time.Second)

}

func testSendMessageToChangingDestination(t *testing.T) {

	// Create test server with the echo handler
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	//	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		echo(w, r)
	//	}))
	//
	//	url := "127.0.0.1:8099"
	//	l, err := net.Listen("tcp", url)
	//
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	s.Listener.Close()
	//	s.Listener = l
	//	s.Start()
	//	defer s.Close()

	// Create test server with the echo handler.
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shout(w, r)
	}))

	defer s2.Close()

	closed := make(chan struct{})

	mh := agg.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	id := "rule0"
	stream := "/stream/large"
	destination := "ws" + strings.TrimPrefix(s.URL, "http")

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	reply := make(chan hub.Message)

	c := &hub.Client{Hub: mh.Hub, Name: "a", Topic: stream, Send: reply}

	mh.Register <- c

	time.Sleep(time.Millisecond)

	payload := []byte("test message")

	mh.Broadcast <- hub.Message{Data: payload, Type: websocket.TextMessage, Sender: *c, Sent: time.Now()}

	msg := <-reply

	if bytes.Compare(msg.Data, payload) != 0 {
		t.Error("Got wrong message")
	}

	destination = "ws" + strings.TrimPrefix(s2.URL, "http")

	r = &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	mh.Register <- c

	time.Sleep(time.Millisecond)

	mh.Broadcast <- hub.Message{Data: payload, Type: websocket.TextMessage, Sender: *c, Sent: time.Now()}

	msg = <-reply

	if bytes.Compare(msg.Data, []byte("TEST MESSAGE")) != 0 {
		t.Error("Did not change server")
	}

	close(closed)
}

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func report(w http.ResponseWriter, r *http.Request, msgChan chan reconws.WsMessage) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		msgChan <- reconws.WsMessage{Data: message, Type: mt}
	}
}

func shout(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		message = []byte(strings.ToUpper(string(message)))
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func suppressLog() {
	var ignore bytes.Buffer
	logignore := bufio.NewWriter(&ignore)
	log.SetOutput(logignore)
}

func displayLog() {
	log.SetOutput(os.Stdout)
}
