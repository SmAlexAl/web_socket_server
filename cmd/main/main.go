package main

import (
	"database/sql"
	"encoding/json"
	"github.com/SmAlexAl/web_socket_server/internal/connection/mysql"
	"github.com/SmAlexAl/web_socket_server/pkg/chat"
	"github.com/SmAlexAl/web_socket_server/pkg/service/JwtService"
	"github.com/gobwas/ws"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

var (
	mainChat  *chat.Chat
	mysqlConn *sql.DB
)

func main() {
	http.HandleFunc("/chat/ws", wsHandler)
	http.HandleFunc("/token", getTokenHanlder)

	godotenv.Load()
	mysqlConn = mysql.Open()
	mainChat = chat.NewMainChat()

	//if err != nil {
	//	fmt.Println(err)
	//}

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
	if err != nil {
		log.Println(err)
	}
	user := mainChat.AddUser(conn)

	user.Reader(mysqlConn)

	chat.Remove(user)

	if err != nil {
		return
	}
}
