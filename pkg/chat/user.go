package chat

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/SmAlexAl/web_socket_server/pkg/service/JwtService"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"sync"
)

type User struct {
	io   sync.Mutex
	conn net.Conn

	id        uint
	name      string
	profileId string
	chat      *Chat
	clanChat  *Chat
}

func (u *User) writeRaw(p []byte) error {
	//u.io.Lock()
	//defer u.io.Unlock()
	_, err := u.conn.Write(p)

	return err
}

func (u *User) Reader(mysqlConn *sql.DB) {
	for {
		requestData, err := u.ReadRequest()

		valid, profileDto := JwtService.ParseToken(requestData.Token, mysqlConn)

		if valid != true {
			response := JwtService.ErrorResponse{
				Message: "Invalid token",
				Code:    10001,
			}
			u.tokenErrorResponse(response)
		} else {
			u.init(*profileDto)

			if err != nil {
				log.Println(err)
				return
			}
			switch requestData.Type {
			case "message":
				requestData.Params["time"] = timestamp()

				err = u.chat.Broadcast("publish", requestData.Params)
				if err != nil {
					log.Println(err)
					return
				}
			case "message_clan":
				requestData.Params["time"] = timestamp()

				err = u.clanChat.Broadcast("publish", requestData.Params)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func (u *User) tokenErrorResponse(response JwtService.ErrorResponse) {
	var buf bytes.Buffer

	w := wsutil.NewWriter(&buf, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	encoder.Encode(response)

	u.writeRaw(buf.Bytes())
}

func (u *User) init(dto JwtService.ProfileDto) {
	if dto.ClanId.Valid != false && u.clanChat.id != dto.ClanId.String {
		addClanChat(u, dto.ClanId.String)
	}

	u.profileId = dto.ProfileId.String
	u.name = dto.ProfileName.String
}

//под вопросом, не уверен что надо
func (u *User) connected(params Object) {
	if params["clanId"] != nil {
		addClanChat(u, params["clanId"].(string))
	}

	if params["name"] != nil {
		u.name = params["name"].(string)
	}

	if params["profileId"] != nil {
		u.profileId = params["profileId"].(string)
	}
}

// readRequests reads json-rpc request from connection.
// It takes io mutex.
func (u *User) ReadRequest() (*Request, error) {
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
