package main

import (
	"net"
	"sync"
)

// Chat contains logic of user interaction.
type Chat struct {
	mu       sync.RWMutex
	seq      uint
	userList []*User
	ns       map[string]*User

	out chan []byte
}

type User struct {
	io   sync.Mutex
	conn net.Conn

	id   uint
	name string
	chat *Chat
}

type Object map[string]interface{}

type Request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params Object `json:"params"`
}

func NewChat() *Chat {
	chat := &Chat{
		ns:  make(map[string]*User),
		out: make(chan []byte, 1),
	}

	go chat.writer()

	return chat
}

// writer writes broadcast messages from chat.out channel.
func (c *Chat) writer() {
	for bts := range c.out {
		c.mu.RLock()
		us := c.userList
		c.mu.RUnlock()

		for _, u := range us {
			u := u // For closure.
			_ = u.writeRaw(bts)

			//c.pool.Schedule(func() {
			//	u.writeRaw(bts)
			//})
		}
	}
}

func (u *User) writeRaw(p []byte) error {
	//u.io.Lock()
	//defer u.io.Unlock()
	_, err := u.conn.Write(p)

	return err
}

func (chat *Chat) addUser(conn net.Conn) *User {
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
