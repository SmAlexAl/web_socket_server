package main

import (
	"database/sql"
	"encoding/json"
	"github.com/SmAlexAl/web_socket_server/internal/connection/mysql"
	"github.com/SmAlexAl/web_socket_server/pkg/chat"
	"github.com/SmAlexAl/web_socket_server/pkg/service/JwtService"
	"github.com/davecgh/go-spew/spew"
	"github.com/gobwas/ws"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	mainChat  *chat.Chat
	mysqlConn *sql.DB
	file      *os.File
)

func main() {
	var err error
	http.HandleFunc("/chat/ws", wsHandler)
	http.HandleFunc("/token", getTokenHanlder)

	spew.Dump("handle ok")
	godotenv.Load()
	mysqlConn, err = mysql.Open()

	if err != nil {
		file.WriteString(err.Error() + "\n")
		panic(err)
	}

	spew.Dump("mysql ok")

	mainChat = chat.NewMainChat()
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	file, err = os.Create("errorWebSocket_" + timeStr)

	spew.Dump("file ok")
	if err != nil {
		log.Println(err)
	}
	defer func() {
		mysqlConn.Close()
		file.Close()
	}()

	panic(http.ListenAndServe(":8080", nil))
}

func getTokenHanlder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)

	tokenData := &JwtService.TokenData{}
	if err := decoder.Decode(tokenData); err != nil {
		log.Println(err)
	}

	token, _ := JwtService.GenerateToken(tokenData)

	responce := JwtService.TokenResponse{
		Token: token,
	}
	json.NewEncoder(w).Encode(responce)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	spew.Dump("socket connection ok")
	if err != nil {
		spew.Dump(err)
		log.Println(err)
	}
	user := mainChat.AddUser(conn)

	err = user.Reader(mysqlConn)

	if err != nil {
		response := chat.ErrorResponse{
			Code:    500,
			Message: err.Error(),
		}
		user.ErrorResponse(response)

		file.WriteString(err.Error() + "\n")
	}

	chat.Remove(user)

	if err != nil {
		return
	}
}
