package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"encoding/json"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v2/pkg/directorysync"
	"github.com/workos/workos-go/v2/pkg/webhooks"
)

type Directory struct {
	ID string
	Directory string
}

type Users struct {
	Users string
}

// A helper function for template
func mod(i, j int) bool {
	return i%j ==0
}

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

func handleDirectory(w http.ResponseWriter, r *http.Request) {


	tmpl := template.Must(template.ParseFiles("./static/directory.html"))
	id := r.URL.Query().Get("id")

	dir, err := directorysync.GetDirectory(
		context.Background(),
		directorysync.GetDirectoryOpts{
			Directory: id,
		},
	)

	if err != nil {
		log.Printf("get directory failed: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	
	b, err := json.MarshalIndent(dir, "", "    ")
	if err != nil {
		log.Printf("encoding directory failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// Stringify the directory details
	newstring := string(b)
	data := Directory{dir.ID, newstring}

	// Render the template with the directory 
	tmpl.Execute(w, data)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("users.html").Funcs(template.FuncMap{"mod": mod}).ParseFiles("./static/users.html")
	id := r.URL.Query().Get("id")

	if err != nil {
		log.Printf("template creation: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	
	// Get Users
	list, err := directorysync.ListUsers(
		context.Background(),
		directorysync.ListUsersOpts{
			Directory: id,
		},
	)

	if err != nil {
		log.Printf("get users failed: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	data := directorysync.ListUsersResponse{list.Data, list.ListMetadata}

	// Render the template with the users as data
	tmpl.Execute(w, data)
}


func handleGroups(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("groups.html").Funcs(template.FuncMap{"mod": mod}).ParseFiles("./static/groups.html")
	id := r.URL.Query().Get("id")

	if err != nil {
		log.Printf("template creation: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	
	// Get Groups
	list, err := directorysync.ListGroups(
		context.Background(),
		directorysync.ListGroupsOpts{
			Directory: id,
		},
	)

	if err != nil {
		log.Printf("get groups failed: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	data := directorysync.ListGroupsResponse{list.Data, list.ListMetadata}
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


	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Directory, "directory", os.Getenv("WORKOS_DIRECTORY_ID"), "The WorkOS directory id.")
	flag.Parse()

	log.Printf("launching directory sync demo with configuration: %+v", conf)

	// Configure the WorkOS directory sync SDK:
	directorysync.SetAPIKey(conf.APIKey)

	// Handle  webhooks
	http.HandleFunc("/webhooks", handleWebhook)

	http.HandleFunc("/", handleDirectories)
	styles := http.FileServer(http.Dir("./static/stylesheets"))
    http.Handle("/styles/", http.StripPrefix("/styles/", styles))

	images := http.FileServer(http.Dir("./static/images"))
    http.Handle("/images/", http.StripPrefix("/images/", images))

	http.HandleFunc("/directory", handleDirectory)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/groups", handleGroups)

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
