package main

import (
	"forum/data"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// GET /login
// Show the login page
func login(writer http.ResponseWriter, request *http.Request) {

	generateHTML(writer, request, nil, "login.layout", "public.navbar", "login")
}

// GET /signup
// Show the signup page
func signup(writer http.ResponseWriter, request *http.Request) {

	generateHTML(writer, request, nil, "login.layout", "public.navbar", "signup")
}

// POST /signupEncrypt(user.Password),
func signupAccount(writer http.ResponseWriter, request *http.Request) {

	err := request.ParseForm()
	if err != nil {
		danger(err, "Cannot parse form")
		error_message(writer, request, "500 Internal Server Error", 500)
	}
	user := data.User{
		Name:     request.PostFormValue("name"),
		Email:    request.PostFormValue("email"),
		Password: "", //request.PostFormValue("password")
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(request.PostFormValue("password")), bcrypt.MinCost)
	user.Password = string(pass)
	if err := user.Create(); err != nil {
		error_message(writer, request, "The email is already taken.", 302)
		danger(err, "Cannot create user")
		return
	}
	http.Redirect(writer, request, "/login", 302)
}

// POST /authenticate
// Authenticate the user given the email and password
func authenticate(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	user, err := data.UserByEmail(request.PostFormValue("email"))
	if err != nil {
		error_message(writer, request, "400 Bad request", 400)
		danger(err, "Cannot find user")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.PostFormValue("password")))
	if err == nil {
		session, err := user.Session()
		if session.Uuid != "" {
			err = session.DeleteByUUID()
		}

		session, err = user.CreateSession()
		if err != nil {
			danger(err, "Cannot create session")
			error_message(writer, request, "500 Internal Server Error", 500)
		}
		cookie := http.Cookie{
			Name:     "_cookie",
			Value:    session.Uuid,
			HttpOnly: true,
		}
		http.SetCookie(writer, &cookie)
		http.Redirect(writer, request, "/", 302)
	} else {
		error_message(writer, request, "Invalid email or password.", 302)
		// http.Redirect(writer, request, "/login", 302)
	}

}

// GET /logout
// Logs the user out
func logout(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("_cookie")
	http.SetCookie(writer, &http.Cookie{
		Name:  "_cookie",
		Value: "",
	})

	if err != http.ErrNoCookie {
		warning(err, "Failed to get cookie")
		session := data.Session{Uuid: cookie.Value}
		session.DeleteByUUID()
	}
	http.Redirect(writer, request, "/", 302)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		generateHTML(w, r, "404 NOT FOUND", "layout", "public.navbar", "error")
	}
}
