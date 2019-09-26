package rwc

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/hub"
)

func init() {

	log.SetLevel(log.PanicLevel)

}

func TestInstantiateHub(t *testing.T) {

	mh := hub.New()

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

	mh := hub.New()
	h := New(mh)
	closed := make(chan struct{})
	go h.Run(closed)

	id := "rule0"
	stream := "/stream/large"
	destination := "ws://localhost:8081"

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
	close(closed)
}

func TestAddRules(t *testing.T) {

	closed := make(chan struct{})

	mh := hub.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	id := "rule0"
	stream := "/stream/large"
	destination := "ws://localhost:8081"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id2 := "rule1"
	stream2 := "/stream/medium"
	destination2 := "ws://localhost:8082"

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
	close(closed)
}

func TestAddDupeRule(t *testing.T) {

	closed := make(chan struct{})

	mh := hub.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	id := "rule0"
	stream := "/stream/large"
	destination := "ws://localhost:8082"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id = "rule0"
	stream = "/stream/medium"
	destination = "ws://localhost:8082"

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

	close(closed)
}

func TestDeleteRule(t *testing.T) {

	closed := make(chan struct{})

	mh := hub.New()
	go mh.Run(closed)

	h := New(mh)
	go h.Run(closed)

	id := "rule0"
	stream := "/stream/large"
	destination := "ws://localhost:8081"

	r := &Rule{Id: id,
		Stream:      stream,
		Destination: destination}

	h.Add <- *r

	time.Sleep(time.Millisecond)

	id2 := "rule1"
	stream2 := "/stream/medium"
	destination2 := "ws://localhost:8082"

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

	close(closed)
}

func TestSendMessage(t *testing.T) {

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo(w, r)
	}))
	defer s.Close()

	closed := make(chan struct{})

	mh := hub.New()
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

	c := &hub.Client{Hub: mh, Name: "a", Topic: stream, Send: reply}

	mh.Register <- c

	time.Sleep(time.Millisecond)

	payload := []byte("test message")

	mh.Broadcast <- hub.Message{Data: payload, Type: websocket.TextMessage, Sender: *c, Sent: time.Now()}

	msg := <-reply

	if bytes.Compare(msg.Data, payload) != 0 {
		t.Error("Got wrong message")
	}

	close(closed)

}

func TestSendMessageToChangingDestination(t *testing.T) {

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

	mh := hub.New()
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

	c := &hub.Client{Hub: mh, Name: "a", Topic: stream, Send: reply}

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
