package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"

	"github.com/gorilla/sessions"
	"github.com/workos/workos-go/pkg/auditlogs"
	"github.com/workos/workos-go/pkg/organizations"
)

var router = http.NewServeMux()
var key = []byte("super-secret-key")
var store = sessions.NewCookieStore(key)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

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
		fmt.Println("problem with response %s'", err)
	}

	session, _ := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["org_id"] = org

	if err := session.Save(r, w); err != nil {
		log.Panic("problem saving cookie %s", err)
	}

	this_response := Organization{response.Name, response.ID}
	tmpl := template.Must(template.ParseFiles("./static/send_events.html"))
	if err := tmpl.Execute(w, this_response); err != nil {
		log.Panic(err)
	}

	auditerr := auditlogs.CreateEvent(context.Background(), auditlogs.AuditLogEventOpts{
		Organization: org,
		Event: auditlogs.Event{
			Action:     "user.organization_set",
			OccurredAt: time.Now(),
			Actor: auditlogs.Actor{
				Type: "user",
				Id:   "user_01GBNJC3MX9ZZJW1FSTF4C5938",
			},
			Targets: []auditlogs.Target{
				{
					Type: "team",
					Id:   "team_01GBNJD4MKHVKJGEWK42JNMBGS",
				},
			},
			Context: auditlogs.Context{
				Location:  "123.123.123.123",
				UserAgent: "Chrome/104.0.0.0",
			},
		},
	},
	)

	if auditerr != nil {
		fmt.Println(err)
	}

}

func sendEvent(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "session-name")
	organization := session.Values["org_id"]

	fmt.Println("organization is %s", organization)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkOrg(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/", http.StatusSeeOther)
}

func main() {
	auditlogs.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	router.HandleFunc("/", checkOrg)
	router.HandleFunc("/send-events", sendEvents)
	router.HandleFunc("/send-event", sendEvent)

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	//Action title: "user.signed_in" | Target type: "team"
	//Action title: "user.logged_out" | Target type: "team"
	//Action title: "user.organization_deleted" | Target type: "team"
	//Action title: "user.connection_deleted" | Target type: "team"

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Panic(err)
	}
}
