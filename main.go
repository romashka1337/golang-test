package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"test-proj/auth"

	"github.com/gorilla/websocket"
)

type ChatMsg struct {
	User    string            `json::"user"`
	Text    string            `json::"text"`
	Headers map[string]string `json:"HEADERS"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func checkToken(authHdr string) (*auth.Cred, error) {
	authStr := strings.ReplaceAll(authHdr, "Bearer ", "")
	cred, err := auth.ValidateToken(authStr)
	return cred, err
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)
		authHdr := r.Header["Authorization"]
		if len(authHdr) == 0 {
			w.WriteHeader(401)
			return
		}
		r.Form = url.Values{}
		cred, err := checkToken(authHdr[0])
		r.Form.Set("auth", cred.UserId)
		if err != nil {
			w.WriteHeader(401)
		}
		return
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	tkn := auth.GenerateToken(r.FormValue("name"))
	w.Header().Add("HX-Redirect", "/home?tkn="+tkn)
	w.Write([]byte(tkn))
}

func profile(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<div>name: " + r.FormValue("auth") + "</div>"))
}

type Client struct {
	messages chan string
}

var clients = make(map[string]Client)

func socket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	cred, err := checkToken(r.URL.Query().Get("tkn"))
	clients[cred.UserId] = Client{messages: make(chan string)}
	defer func() {
		clients[cred.UserId].messages <- "close"
		delete(clients, cred.UserId)
	}()
	if _, ok := clients[cred.UserId]; ok {
		go func() {
			defer close(clients[cred.UserId].messages)
			for {
				select {
				case msg := <-clients[cred.UserId].messages:
					if msg == "close" {
						return
					}
					if err := conn.WriteMessage(1, []byte(msg)); err != nil {
						log.Println(err)
						return
					}
				}
			}
		}()
	}
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		sct := &ChatMsg{}
		json.Unmarshal(p, sct)
		if sct.Text == "" {
			continue
		}
		if err != nil {
			w.WriteHeader(401)
		}

		div := "<div id=\"msg\" hx-swap-oob=\"beforeend\">"
		div += cred.UserId + ":\t"
		div += sct.Text
		div += "<br>"
		div += "</div>"
		for _, v := range clients {
			v.messages <- div
		}
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/login.html")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Replace-Url", "/home?tkn=")
	w.Write([]byte(Home(r.URL.Query().Get("tkn"))))
}

func main() {
	http.HandleFunc("/", loginPage)
	http.HandleFunc("/home", homePage)
	http.HandleFunc("/socket", socket)
	http.HandleFunc("/login", login)
	http.Handle("/user/profile", authMiddleware(http.HandlerFunc(profile)))

	http.ListenAndServe(":8080", nil)
}
