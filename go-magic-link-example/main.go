package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v2/pkg/passwordless"
	"github.com/workos/workos-go/v2/pkg/sso"
)

type Profile struct {
	Email   string
	Session string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		Addr     string
		APIKey   string
		ClientID string
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.ClientID, "client-id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS client id.")

	log.Printf("launching magic link demo with configuration: %+v", conf)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	sso.Configure(conf.APIKey, conf.ClientID)

	http.HandleFunc("/passwordless-auth", func(w http.ResponseWriter, r *http.Request) {
		passwordless.SetAPIKey(conf.APIKey)
		if err := r.ParseForm(); err != nil {
			log.Panic(err)
		}

		email := r.Form["email"][0]

		session, err := passwordless.CreateSession(context.Background(), passwordless.CreateSessionOpts{
			Email: email,
			Type:  passwordless.MagicLink,
		})

		if err != nil {
			fmt.Println(err)
		}

		err = passwordless.SendSession(context.Background(), passwordless.SendSessionOpts{
			SessionID: session.ID,
		})

		if err != nil {
			fmt.Println(err)
		}

		this_profile := Profile{email, session.Link}
		tmpl := template.Must(template.ParseFiles("./static/serve_magic_link.html"))
		if err := tmpl.Execute(w, this_profile); err != nil {
			log.Panic(err)
		}
	})

	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		profileAndToken, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOpts{
			Code: r.URL.Query().Get("code"),
		})

		if err != nil {
			fmt.Println(err)
		}

		// Use the information in `profile` for further business logic.
		profile := profileAndToken.Profile

		Raw_profile, err := json.MarshalIndent(profile, "", "    ")
		if err != nil {
			log.Println(err)
			return
		}

		tmpl := template.Must(template.ParseFiles("./static/success.html"))
		if err := tmpl.Execute(w, string(Raw_profile)); err != nil {
			log.Panic(err)
		}
	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
