package main

import (
	"net/http"
	"strconv"

	"forum/data"
)

// GET /err?msg=
// shows the error message page
func err(writer http.ResponseWriter, request *http.Request) {
	vals := request.URL.Query()
	_, err := session(writer, request)
	if err != nil {
		generateHTML(writer, request, vals.Get("msg"), "layout", "public.navbar", "error")
	} else {
		generateHTML(writer, request, vals.Get("msg"), "layout", "private.navbar", "error")
	}
}

func index(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" && request.URL.Path != "/filter/created" && request.URL.Path != "/filter/liked" && request.URL.Path != "/filter/category" {
		errorHandler(writer, request, http.StatusNotFound)
		return
	}
	err := request.ParseForm()
	if err != nil {
		return
	}
	sortBy := request.PostFormValue("type")
	if sortBy == "created" {
		sess, err := session(writer, request)
		if err != nil {
			http.Redirect(writer, request, "/login", 302)
		} else {
			user, err := sess.User()
			threads, err := user.CreatedThreads()
			if err != nil {
				error_message(writer, request, "400 Bad request", 400)
			} else {
				_, err := session(writer, request)
				if err != nil {
					generateHTML(writer, request, threads, "layout", "public.navbar", "index")
				} else {
					generateHTML(writer, request, threads, "layout", "private.navbar", "index")
				}
			}
		}
	} else if sortBy == "liked" {
		sess, err := session(writer, request)
		if err != nil {
			http.Redirect(writer, request, "/login", 302)
		} else {
			user, err := sess.User()
			threads, err := user.LikedThreads()
			if err != nil {
				error_message(writer, request, "400 Bad request", 400)
			} else {
				_, err := session(writer, request)
				if err != nil {
					generateHTML(writer, request, threads, "layout", "public.navbar", "index")
				} else {
					generateHTML(writer, request, threads, "layout", "private.navbar", "index")
				}
			}
		}
	} else if sortBy == "category" {
		var categories []string
		for i := 1; i <= 5; i++ {
			temp := request.PostFormValue("category" + strconv.Itoa(i))
			if temp != "" {
				categories = append(categories, temp)
			}
		}
		threads, err := data.ThreadsByCategories(categories)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
		} else {
			_, err := session(writer, request)
			if err != nil {
				generateHTML(writer, request, threads, "layout", "public.navbar", "index")
			} else {
				generateHTML(writer, request, threads, "layout", "private.navbar", "index")
			}
		}
	} else {
		threads, err := data.Threads()
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
		} else {
			_, err := session(writer, request)
			if err != nil {
				generateHTML(writer, request, threads, "layout", "public.navbar", "index")
			} else {
				generateHTML(writer, request, threads, "layout", "private.navbar", "index")
			}
		}
	}
}
