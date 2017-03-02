package main

import (
	"net/http"

	session "github.com/huqiangit/negroni_session"

	"github.com/urfave/negroni"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello world"))
	})

	n := negroni.Classic()
	n.Use(session.DefaultSession)
	n.UseHandler(mux)

	http.ListenAndServe(":3003", n)
}
