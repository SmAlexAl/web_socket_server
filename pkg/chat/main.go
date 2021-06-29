package chat

import (
	"time"
)

var clanChatList = map[string]*Chat{}

func NewMainChat() *Chat {
	chat := &Chat{
		id:  "11111",
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

func addClanChat(user *User, clanId string) {
	if clanChatList[clanId] != nil {
		ccl := clanChatList[clanId]

		if user.clanChat != nil {
			user.clanChat.removeUserFromChat(user)
		}

		user.clanChat = ccl

		ccl.mu.Lock()
		{
			ccl.userList = append(ccl.userList, user)
			ccl.ns[user.name] = user
		}
		ccl.mu.Unlock()
	} else {
		ccl := createClanChat(clanId)
		user.clanChat = ccl

		ccl.mu.Lock()
		{
			ccl.userList = append(ccl.userList, user)
			ccl.ns[user.name] = user
		}
		ccl.mu.Unlock()
	}
}

// mutex must be held.
func Remove(user *User) bool {
	mainChat := user.chat.removeUserFromChat(user)

	clanChat := true
	if user.clanChat != nil {
		clanChat = user.clanChat.removeUserFromChat(user)
	}

	return mainChat && clanChat
}

func createClanChat(clanId string) *Chat {
	chat := &Chat{
		id:  clanId,
		ns:  make(map[string]*User),
		out: make(chan []byte, 1),
	}

	clanChatList[clanId] = chat

	go chat.writer()

	return chat
}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
