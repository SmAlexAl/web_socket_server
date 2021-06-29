package chat

import (
	"bytes"
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"sort"
	"sync"
)

// Chat contains logic of user interaction.
// в дальнейшем добавить pool go рутин для ограничения очереди записи
type Chat struct {
	id       string
	mu       sync.RWMutex
	seq      uint
	userList []*User
	ns       map[string]*User

	out chan []byte
}

type Object map[string]interface{}

type Request struct {
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Token  string `json:"token"`
	Params Object `json:"params"`
}

// Broadcast sends message to all alive users.
func (c *Chat) Broadcast(method string, params Object) error {
	var buf bytes.Buffer

	w := wsutil.NewWriter(&buf, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	r := Request{Type: method, Params: params}
	if err := encoder.Encode(r); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	c.out <- buf.Bytes()

	return nil
}

func (chat *Chat) AddUser(conn net.Conn) *User {
	user := &User{
		chat: chat,
		conn: conn,
	}

	chat.mu.Lock()
	{
		user.id = chat.seq

		chat.userList = append(chat.userList, user)
		chat.ns[user.name] = user

		chat.seq++
	}
	chat.mu.Unlock()

	return user
}

func (c *Chat) removeUserFromChat(user *User) bool {
	i := sort.Search(len(c.userList), func(i int) bool {
		return c.userList[i].profileId == user.profileId
	})

	if i >= len(c.userList) {
		panic("chat: inconsistent state")
	}

	if i > 0 {
		i--
	}

	without := make([]*User, len(c.userList)-1)
	copy(without[:i], c.userList[:i])
	copy(without[i:], c.userList[i+1:])
	c.userList = without

	return true
}
