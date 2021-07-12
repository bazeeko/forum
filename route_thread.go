package main

import (
	"fmt"
	"forum/data"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// GET /threads/new
// Show the new thread form page
func newThread(writer http.ResponseWriter, request *http.Request) {
	_, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		generateHTML(writer, request, nil, "layout", "private.navbar", "new.thread")
	}
}

// POST /signup
// Create the user account
func createThread(writer http.ResponseWriter, request *http.Request) {

	imgName := ""
	img, header, err := request.FormFile("img")

	if img != nil {
		defer img.Close()
		if err != nil {
			danger(err, "Cannot upload image.")
			error_message(writer, request, "Cannot upload image.", 500)
			return
		}
		imgData, err := ioutil.ReadAll(img)
		if err != nil {
			danger(err, "Cannot upload image")
			error_message(writer, request, "500 Internal Server Error.", 500)
			return
		}

		if header.Size >= 20971520 {
			danger(err, "File size is more than 20MB.")
			error_message(writer, request, "400 Bad Request. Cannot upload image. Image size should be less than 20MB", 400)
			return
		}

		var magicTable = map[string]string{ //table for checking image type
			"\xff\xd8\xff":      ".jpeg",
			"\x89PNG\r\n\x1a\n": ".png",
			"GIF87a":            ".gif",
			"GIF89a":            ".gif",
		}
		var format string
		for magic, mime := range magicTable {
			if strings.HasPrefix(string(imgData), magic) {
				format = mime
				break
			}
		}
		if format == "" {
			danger(err, "Image is damaged or image type is not supported.")
			error_message(writer, request, "400 Bad Request. Cannot upload image: image is damaged or image type is not supported.", 400)
			return
		}

		imgName = "img-" + data.CreateUUID() + format
		path := filepath.Join("public/images", imgName)
		err = ioutil.WriteFile(path, imgData, 0666)
		if err != nil {
			danger(err, "Cannot upload image.")
			error_message(writer, request, "500 Internal Server Error. Cannot upload image.", 500)
			return
		}
	}

	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		var categories []string
		for i := 1; i <= 5; i++ {
			temp := request.PostFormValue("category" + strconv.Itoa(i))
			if temp != "" {
				categories = append(categories, temp)
			}
		}
		topic := request.PostFormValue("topic")
		if strings.Trim(topic, " ") == "" {
			error_message(writer, request, "Cannot create an empty post", 400)
			return
		}
		_, err = user.CreateThread(topic, categories, imgName)
		if err != nil {
			danger(err, "Cannot create thread.")
			error_message(writer, request, "500 Internal Server Error.", 500)
			return
		}

		http.Redirect(writer, request, "/", 302)
	}
}

// GET /thread/read
// Show the details of the thread, including the posts and the form to write a post
func readThread(writer http.ResponseWriter, request *http.Request) {
	vals := request.URL.Query()
	uuid := vals.Get("id")
	thread, err := data.ThreadByUUID(uuid)
	if err != nil {
		error_message(writer, request, "404 Page Not Found", 404)
	} else {
		_, err := session(writer, request)
		if err != nil {
			generateHTML(writer, request, &thread, "layout", "public.navbar", "public.thread")
		} else {
			generateHTML(writer, request, &thread, "layout", "private.navbar", "private.thread")
		}
	}
}

// POST /thread/post
// Create the post
func postThread(writer http.ResponseWriter, request *http.Request) {
	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		body := request.PostFormValue("body")
		if strings.Trim(body, " ") == "" {
			error_message(writer, request, "Cannot create an empty comment", 400)
			return
		}
		uuid := request.PostFormValue("uuid")
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
			return
		}
		if _, err := user.CreatePost(thread, body); err != nil {
			danger(err, "Cannot create post")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		url := fmt.Sprint("/thread/read?id=", uuid)
		http.Redirect(writer, request, url, 302)
	}
}

func addThreadLike(writer http.ResponseWriter, request *http.Request) {
	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		uuid := request.PostFormValue("uuid")
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
			return
		}
		if err := user.RateThread(thread); err != nil {
			danger(err, "Cannot rate thread")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		url := request.PostFormValue("url")
		http.Redirect(writer, request, url, 302)
	}
}

func addThreadDislike(writer http.ResponseWriter, request *http.Request) {
	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		uuid := request.PostFormValue("uuid")
		thread, err := data.ThreadByUUID(uuid)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
			return
		}
		if err := user.UnrateThread(thread); err != nil {
			danger(err, "Cannot rate thread")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		url := request.PostFormValue("url")
		http.Redirect(writer, request, url, 302)
	}
}

func addPostLike(writer http.ResponseWriter, request *http.Request) {
	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		uuid := request.PostFormValue("uuid")
		post, err := data.PostByUUID(uuid)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
			return
		}
		if err := user.RatePost(post); err != nil {
			danger(err, "Cannot rate post")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		url := request.PostFormValue("url")
		http.Redirect(writer, request, url, 302)
	}
}

func addPostDislike(writer http.ResponseWriter, request *http.Request) {
	sess, err := session(writer, request)
	if err != nil {
		http.Redirect(writer, request, "/login", 302)
	} else {
		err = request.ParseForm()
		if err != nil {
			danger(err, "Cannot parse form")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		user, err := sess.User()
		if err != nil {
			danger(err, "Cannot get user from session")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		uuid := request.PostFormValue("uuid")
		post, err := data.PostByUUID(uuid)
		if err != nil {
			error_message(writer, request, "400 Bad request", 400)
			return
		}
		if err := user.UnratePost(post); err != nil {
			danger(err, "Cannot rate post")
			error_message(writer, request, "500 Internal Server Error", 500)
			return
		}
		url := request.PostFormValue("url")
		http.Redirect(writer, request, url, 302)
	}
}

// func HandleUpload(writer http.ResponseWriter, request *http.Request) {
// 	in, header, err := request.FormFile("file")
// 	if err != nil {
// 		danger(err, "Cannot upload image.")
// 		error_message(writer, request, "Cannot upload image.")
// 	}
// 	data, err := ioutil.ReadAll(in)
// 	if err != nil {
// 		fmt.Println("SASATTT")
// 		return
// 	}
// 	defer in.Close()
// 	//you probably want to make sure header.Filename is unique and
// 	// use filepath.Join to put it somewhere else.
// 	ioutil.WriteFile("image.jpg", data, 0666)
// 	out, err := os.OpenFile(header.Filename, os.O_WRONLY, 0644)
// 	if err != nil {
// 		//handle error
// 	}
// 	defer out.Close()
// 	io.Copy(out, in)
// 	//do other stuff
// }

// func DeteteThread(writer http.ResponseWriter, request *http.Request) {
// 	sess, err := session(writer, request)
// 	if err != nil {
// 		http.Redirect(writer, request, "/login", 400)
// 	} else {
// 		err = request.ParseForm()
// 		if err != nil {
// 			danger(err, "Cannot parse form")
// 			error_message(writer, request, "500 Internal Server Error")
// 			return
// 		}
// 		user, err := sess.User()
// 		if err != nil {
// 			danger(err, "Cannot get user from session")
// 			error_message(writer, request, "500 Internal Server Error")
// 			return
// 		}
// 		var categories []string
// 		for i := 1; i <= 5; i++ {
// 			temp := request.PostFormValue("category" + strconv.Itoa(i))
// 			if temp != "" {
// 				categories = append(categories, temp)
// 			}
// 		}
// 		topic := request.PostFormValue("topic")
// 		if strings.Trim(topic, " ") == "" {
// 			error_message(writer, request, "Cannot create an empty post")
// 			return
// 		}

// 		http.Redirect(writer, request, "/", 302)
// 	}
// }
