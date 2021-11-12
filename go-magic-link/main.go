package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/workos-inc/workos-go/pkg/passwordless"
	"github.com/workos-inc/workos-go/pkg/sso"
)

func main() {
	address := ":8000"
	apiKey := 
	clientID := 

	http.Handle("/", http.FileServer(http.Dir("./static")))

	sso.Configure(apiKey, clientID)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {

		profileAndToken, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOptions{
			Code: r.URL.Query().Get("code"),
		})

		if err != nil {
			fmt.Println(err)
		}

		// Use the information in `profile` for further business logic.
		profile := profileAndToken.Profile
		fmt.Println(profile)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/passwordless-auth", func(w http.ResponseWriter, r *http.Request) {
		passwordless.SetAPIKey(apiKey)
		r.ParseForm()

		email := r.Form["email"][0]

		session, err := passwordless.CreateSession(context.Background(), passwordless.CreateSessionOpts{
			Email: email,
			Type:  passwordless.MagicLink,
		})

		if err != nil {
			fmt.Println(err, "first")
		}

		err = passwordless.SendSession(context.Background(), passwordless.SendSessionOpts{
			ID: session.ID,
		})

		if err != nil {
			fmt.Println(err, "second")
		}
		tmpl := template.Must(template.ParseFiles("./static/serve_magic_link.html"))
		tmpl.Execute(w, email)
	})

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Panic(err)
	}

}
