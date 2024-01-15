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
	Close  bool
}

type Client struct {
	messages chan Message
}

var clients = make(map[string]Client)

func handlemessages(conn *websocket.Conn, user string) {
	for {
		select {
		case msg := <-clients[user].messages:
			if msg.Close {
				return
			}
			tmpl, _ := template.ParseFiles("html/message.html")
			var res bytes.Buffer
			if err := tmpl.Execute(&res, msg); err != nil {
				log.Printf("Ex: %v", err)
				return
			}
			if err := conn.WriteMessage(1, []byte(res.String())); err != nil {
				log.Printf("Wm: %v", err)
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
		clients[userId].messages <- Message{Close: true}
		close(clients[userId].messages)
		delete(clients, userId)
	}()
	go handlemessages(conn, userId)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("RD: %v", err)
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
