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

type Profile struct {
	Email   string
	Session string
}

func main() {
	address := ":8000"
	apiKey := 
	clientID := 

	http.Handle("/", http.FileServer(http.Dir("./static")))

	sso.Configure(apiKey, clientID)

	//http.Handle("/success", http.FileServer(http.Dir("./static/success")))

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {

		sso.Configure(apiKey, clientID)
		profileAndToken, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOptions{
			Code: r.URL.Query().Get("code"),
		})

		if err != nil {
			fmt.Println(err)
		}

		// Use the information in `profile` for further business logic.
		profile := profileAndToken.Profile
		fmt.Println(profile)
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
			fmt.Println(err)
		}

		err = passwordless.SendSession(context.Background(), passwordless.SendSessionOpts{
			ID: session.ID,
		})

		if err != nil {
			fmt.Println(err)
		}

		this_profile := Profile{email, session.Link}
		tmpl := template.Must(template.ParseFiles("./static/serve_magic_link.html"))
		tmpl.Execute(w, this_profile)
	})

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Panic(err)
	}

}
