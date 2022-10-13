package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/auditlogs"
	"github.com/workos/workos-go/pkg/organizations"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

var router = http.NewServeMux()

type Organization struct {
	Name string
	ID   string
}

func sendEvents(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	org := r.FormValue("org")

	organizations.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	response, err := organizations.GetOrganization(
		context.Background(),
		organizations.GetOrganizationOpts{
			Organization: org,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	this_response := Organization{response.Name, response.ID}
	tmpl := template.Must(template.ParseFiles("./static/send_events.html"))
	if err := tmpl.Execute(w, this_response); err != nil {
		log.Panic(err)
	}
}

func main() {
	auditlogs.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/send-events", sendEvents)
	//Action title: "user.signed_in" | Target type: "team"
	//Action title: "user.logged_out" | Target type: "team"
	//Action title: "user.organization_set" | Target type: "team"
	//Action title: "user.organization_deleted" | Target type: "team"
	//Action title: "user.connection_deleted" | Target type: "team"

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Panic(err)
	}
}
