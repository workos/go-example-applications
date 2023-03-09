package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/gorilla/sessions"
	"github.com/workos/workos-go/v2/pkg/auditlogs"
	"github.com/workos/workos-go/v2/pkg/organizations"
	"github.com/workos/workos-go/v2/pkg/portal"
	"github.com/workos/workos-go/v2/pkg/common"
	
)

var router = http.NewServeMux()
var key = []byte("super-secret-key")
var store = sessions.NewCookieStore(key)

type SendEventData struct{
	Name string
	ID   string
	RangeStart string
	RangeEnd string
}

type Organizations struct {
	Data []organizations.Organization
	Metadata common.ListMetadata
	Before string
	After string
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}


// Displays Organizations 
func handleOrganizations(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/index.html"))

	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")

	list, err := organizations.ListOrganizations(
		context.Background(),
		organizations.ListOrganizationsOpts{
			Before: before,
			After: after,
			Limit: 5,
		},
	)

	before = list.ListMetadata.Before
	after = list.ListMetadata.After

	// Get orgs
	if err != nil {
		log.Printf("get organizations failed: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	
	data := Organizations{list.Data, list.ListMetadata, before, after}

	// Render the template with the organizations
	tmpl.Execute(w, data)
}

func getOrg(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	org := r.URL.Query().Get("id")

	response, err := organizations.GetOrganization(
		context.Background(),
		organizations.GetOrganizationOpts{
			Organization: org,
		},
	)

	if err != nil {
		fmt.Println("Problem with response:", err)
	}

	session, _ := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["org_id"] = org
	session.Values["org_name"] = response.Name

	if err := session.Save(r, w); err != nil {
		log.Panic("problem saving cookie:", err)
	}

	auditerr := auditlogs.CreateEvent(context.Background(), auditlogs.CreateEventOpts{
		OrganizationID: org,
		Event: auditlogs.Event{
			Action:     "user.organization_set",
			OccurredAt: time.Now(),
			Actor: auditlogs.Actor{
				Type: "user",
				ID:   "user_01GBNJC3MX9ZZJW1FSTF4C5938",
			},
			Targets: []auditlogs.Target{
				{
					Type: "team",
					ID:   "team_01GBNJD4MKHVKJGEWK42JNMBGS",
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
		fmt.Println(auditerr)
	}

	http.Redirect(w, r, "/send-events", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Revoke users authentication
	session.Values["org_id"] = nil
	session.Values["org_name"] = nil
	session.Values["export_id"] = nil

	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Executes the send_events template
func sendEvents(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	tmpl := template.Must(template.ParseFiles("./static/send_events.html"))
	currentTime := time.Now()
	rangeStart := currentTime.AddDate(0, 0, -30)

	data := SendEventData{session.Values["org_name"].(string), session.Values["org_id"].(string), rangeStart.Format(time.RFC3339), currentTime.Format(time.RFC3339)}
	
	if err := tmpl.Execute(w, data); err != nil {
		log.Panic(err)
	}
}

// Submits an event on form submission
func sendEvent(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	eventVersion, err := strconv.Atoi(r.FormValue("event-version"))
	if err != nil {
		fmt.Println("Couldn't convert version to string", err)
	}
	actorName := r.FormValue("actor-name")
	actorType := r.FormValue("actor-type")
	targetName := r.FormValue("target-name")
	targetType := r.FormValue("target-type")

	err = auditlogs.CreateEvent(context.Background(), auditlogs.CreateEventOpts{
		OrganizationID: session.Values["org_id"].(string),
		Event: auditlogs.Event{
			Action:     "user.organization_deleted",
			OccurredAt: time.Now(),
			Version:    eventVersion,
			Actor: auditlogs.Actor{
				ID:   "user_TF4C5938",
				Type: actorType,
				Name: actorName,
			},
			Targets: []auditlogs.Target{
				{
					Type: targetType,
					ID:   "user_98432YHF",
					Name: targetName,
				},
			},
			Context: auditlogs.Context{
				Location:  "1.1.1.1",
				UserAgent: "Chrome/104.0.0.0",
			},
		},
	})

	if err != nil {
		fmt.Println("Couldn't send event:", err)
	}

	http.Redirect(w, r, "/send-events", http.StatusSeeOther)
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	before := r.URL.Query().Get("before")
	after := r.URL.Query().Get("after")

	// if pagination parameters not present, check for a session
	if before == "" && after == "" {
		if val, ok := session.Values["org_id"].(string); ok {
			// if val is a string
			switch val {
			case "":
				http.Redirect(w, r, "/index", http.StatusSeeOther)
			default:
				http.Redirect(w, r, "/send-events", http.StatusSeeOther)
			}
		} else {
			// if val is not a string type
			http.Redirect(w, r, "/index", http.StatusSeeOther)
		}
	}
}

// Generates or exports CSV by form event 
func exportEvents(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	eventType  := r.FormValue("event")

	if eventType == "access_csv" {
		export, err := auditlogs.GetExport(context.Background(), auditlogs.GetExportOpts{
			ExportID: session.Values["export_id"].(string),
		})
	
		if err != nil {
			fmt.Println("Error exporting events:", err)
		}
	
		http.Redirect(w, r, export.URL, http.StatusSeeOther)
	} else {
		rangeStart := r.FormValue("range-start")
		rangeEnd := r.FormValue("range-end")
		actions := r.FormValue("actions")
		actors := r.FormValue("actors")
		targets := r.FormValue("targets")
		
		export, err := auditlogs.CreateExport(context.Background(), auditlogs.CreateExportOpts{
			OrganizationID: session.Values["org_id"].(string),
			RangeStart:     rangeStart,
			RangeEnd:       rangeEnd,
			Actions:        []string{actions},
			Actors:         []string{actors},
			Targets:        []string{targets},
		})

		if err != nil {
			fmt.Println("Error creating export:", err)
		}

		
		session.Values["export_id"] = export.ID

		if err := session.Save(r, w); err != nil {
			log.Panic("problem saving cookie", err)
		}

		http.Redirect(w, r, "/send-events", http.StatusSeeOther)
	}
}

// Generates an Admin Portal Link by Intent
func events(w http.ResponseWriter, r *http.Request){
	intent := r.URL.Query().Get("intent")
	session, _ := store.Get(r, "session-name")
	
	link, err := portal.GenerateLink(
		context.Background(),
		portal.GenerateLinkOpts{
			Organization: session.Values["org_id"].(string),
			Intent:       portal.GenerateLinkIntent(intent),
		},
	)

	if err != nil {
		fmt.Println("Error generating portal link:", err)
	}

	http.Redirect(w, r, link, http.StatusFound)

}

func main() {
	auditlogs.SetAPIKey(os.Getenv("WORKOS_API_KEY"))
	organizations.SetAPIKey(os.Getenv("WORKOS_API_KEY"))
	portal.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	log.Printf("launching audit log demo")

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", sessionHandler)
	router.HandleFunc("/index", handleOrganizations)
	router.HandleFunc("/get-org", getOrg)
	router.HandleFunc("/events", events)
	router.HandleFunc("/send-events", sendEvents)
	router.HandleFunc("/send-event", sendEvent)
	router.HandleFunc("/export-events", exportEvents)
	router.HandleFunc("/logout", logout)


	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Panic(err)
	}
}
