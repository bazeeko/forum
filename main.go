package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("CodeForum 01 started at", config.Address)
	// p("CodeForum", version(), "started at", config.Address)

	// handle static assets
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir(config.Static))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	// all route patterns matched here
	// route handler functions defined in other files

	// index
	mux.HandleFunc("/", index)
	mux.HandleFunc("/filter/created", index)
	mux.HandleFunc("/filter/liked", index)
	mux.HandleFunc("/filter/category", index)
	// error
	mux.HandleFunc("/err", err)

	// defined in route_auth.go
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/signup", signup)
	mux.HandleFunc("/signup_account", signupAccount)
	mux.HandleFunc("/authenticate", authenticate)

	// defined in route_thread.go
	mux.HandleFunc("/thread/new", newThread)
	mux.HandleFunc("/thread/create", createThread)
	mux.HandleFunc("/thread/post", postThread)
	mux.HandleFunc("/thread/read", readThread)
	mux.HandleFunc("/thread/like", addThreadLike)
	mux.HandleFunc("/thread/dislike", addThreadDislike)
	mux.HandleFunc("/post/like", addPostLike)
	mux.HandleFunc("/post/dislike", addPostDislike)
	// mux.HandleFunc("/image/upload", HandleUpload)

	// starting up the server
	server := &http.Server{
		Addr:           config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
}
