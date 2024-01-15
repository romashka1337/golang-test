package main

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"test-proj/auth"
	"test-proj/chat"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)
		authHdr := r.Header["Authorization"]
		if len(authHdr) == 0 {
			authHdr = []string{r.URL.Query().Get("tkn")}
		}
		if len(authHdr) == 0 {
			w.WriteHeader(401)
			return
		}
		r.Form = url.Values{}
		authStr := strings.ReplaceAll(authHdr[0], "Bearer ", "")
		cred, err := auth.ValidateToken(authStr)
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

func loginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/login.html")
}

type Token struct {
	Token string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Replace-Url", "/")
	tmpl := template.Must(template.ParseFiles("html/home.html"))
	data := Token{Token: r.URL.Query().Get("tkn")}
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", loginPage)
	http.HandleFunc("/home", homePage)
	http.Handle("/socket", authMiddleware(http.HandlerFunc(chat.Socket)))
	http.HandleFunc("/login", login)
	http.Handle("/user/profile", authMiddleware(http.HandlerFunc(profile)))

	http.ListenAndServe(":8080", nil)
}
