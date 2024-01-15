package chat

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	UserId string
	Text   string
}

type Client struct {
	messages chan Message
}

var clients = make(map[string]Client)

func handlemessages(conn *websocket.Conn, user string) {
	defer close(clients[user].messages)
	for {
		select {
		case msg := <-clients[user].messages:
			if msg.Text == "close" {
				return
			}
			tmpl, _ := template.ParseFiles("html/message.html")
			var res bytes.Buffer
			if err := tmpl.Execute(&res, msg); err != nil {
				log.Println(err)
				return
			}
			if err := conn.WriteMessage(1, []byte(res.String())); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func Socket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	userId := r.FormValue("auth")
	clients[userId] = Client{messages: make(chan Message)}
	defer func() {
		clients[userId].messages <- Message{Text: "close"}
		delete(clients, userId)
	}()
	go handlemessages(conn, userId)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		sct := &Message{}
		json.Unmarshal(p, sct)
		if sct.Text == "" {
			continue
		}
		if err != nil {
			w.WriteHeader(401)
		}

		for _, v := range clients {
			v.messages <- Message{UserId: userId, Text: sct.Text}
		}
	}
}
