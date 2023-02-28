package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v2/pkg/directorysync"
	"github.com/workos/workos-go/v2/pkg/webhooks"
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	bodyString := string(body)
	hooks := webhooks.NewClient(os.Getenv("WORKOS_WEBHOOK_SECRET"))
	hook, err := hooks.ValidatePayload(r.Header.Get("Workos-Signature"), bodyString)
	fmt.Println("WORKOS WEBHOOKS")
	if err != nil {
		fmt.Println("Errors found:")
		fmt.Println(err)
	} else {
		fmt.Println("Webhook Succesfully Validated:")
		fmt.Println(hook)
	}
}

func handleDirectories(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/index.html"))

	// Get list of directories
	list, err := directorysync.ListDirectories(
		context.Background(),
		directorysync.ListDirectoriesOpts{},
	)

	if err != nil {
		log.Printf("get list of directories failed: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	
	data := directorysync.ListDirectoriesResponse{list.Data, list.ListMetadata}
	// Render the template with the users as data
	tmpl.Execute(w, data)
}

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

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Directory, "directory", os.Getenv("WORKOS_DIRECTORY_ID"), "The WorkOS directory id.")
	flag.Parse()

	log.Printf("launching directory sync demo with configuration: %+v", conf)


	// Configure the WorkOS directory sync SDK:
	directorysync.SetAPIKey(conf.APIKey)

	// Handle users redirect:
	tmpl2 := template.Must(template.ParseFiles("./static/users.html"))
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
		tmpl2.Execute(w, data)
	})

	// Handle  webhooks
	http.HandleFunc("/webhooks", handleWebhook)

	http.HandleFunc("/", handleDirectories)
	styles := http.FileServer(http.Dir("./static/stylesheets"))
    http.Handle("/styles/", http.StripPrefix("/styles/", styles))

	

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
