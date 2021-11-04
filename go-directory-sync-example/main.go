package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos-inc/workos-go/pkg/directorysync"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		Addr      string
		APIKey    string
		ProjectID string
		Directory string
	}

	type Users struct {
		Users string
	}

	flag.StringVar(&conf.Addr, "addr", ":3042", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Directory, "directory", os.Getenv("WORKOS_DIRECTORY_ID"), "The WorkOS directory id.")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Configure the WorkOS directory sync SDK:
	directorysync.SetAPIKey(conf.APIKey)

	// Handle users redirect:
	tmpl := template.Must(template.ParseFiles("./static/users.html"))
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {

		// Retrieving user profile:
		users, err := directorysync.ListUsers(context.Background(), directorysync.ListUsersOpts{
			Directory: conf.Directory,
		})
		if err != nil {
			log.Printf("get list users failed: %s", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// Display Lists:
		b, err := json.MarshalIndent(users, "", "    ")
		if err != nil {
			log.Printf("encoding list users failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		// Stringify the returned users
		stringB := string(b)
		data := Users{stringB}

		// Render the template with the users as data
		tmpl.Execute(w, data)
	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
