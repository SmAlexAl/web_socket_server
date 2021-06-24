package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net/http"
	"time"
)

var mainChat *Chat

//var poller netpoll.Poller

func main() {
	http.HandleFunc("/chat/ws", wsHandler)
	//http.HandleFunc("/", rootHandler)

	mainChat = NewChat()

	//poller, _ = netpoll.New(nil)

	//if err != nil {
	//	fmt.Println(err)
	//}

	panic(http.ListenAndServe(":8080", nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Println(err)
	}

	//acceptDesc := netpoll.Must(netpoll.Handle(
	//	conn, netpoll.EventRead|netpoll.EventOneShot,
	//))
	//err = web.WriteMessage(1, []byte("Hi Client!"))
	// helpful log statement to show connections
	fmt.Println("Client Connected")

	user := mainChat.addUser(conn)

	user.reader()
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (u *User) reader() {
	for {
		req, err := u.readRequest()

		if err != nil {
			log.Println(err)
			return
		}

		req.Params["time"] = timestamp()

		u.chat.Broadcast("publish", req.Params)
		spew.Dump(req)
	}
}

// readRequests reads json-rpc request from connection.
// It takes io mutex.
func (u *User) readRequest() (*Request, error) {
	u.io.Lock()
	defer u.io.Unlock()

	h, r, err := wsutil.NextReader(u.conn, ws.StateServerSide)
	if err != nil {
		return nil, err
	}
	if h.OpCode.IsControl() {
		return nil, wsutil.ControlFrameHandler(u.conn, ws.StateServerSide)(h, r)
	}
	req := &Request{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Broadcast sends message to all alive users.
func (c *Chat) Broadcast(method string, params Object) error {
	var buf bytes.Buffer

	w := wsutil.NewWriter(&buf, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	r := Request{Method: method, Params: params}
	if err := encoder.Encode(r); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	c.out <- buf.Bytes()

	return nil
}
