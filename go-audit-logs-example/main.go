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

func getOrg(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println("problem with response:", err)
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
		fmt.Println(auditerr)
	}

	http.Redirect(w, r, "/send-events", http.StatusSeeOther)
}

func sendEvents(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	organization := Organization{session.Values["org_name"].(string), session.Values["org_id"].(string)}
	tmpl := template.Must(template.ParseFiles("./static/send_events.html"))
	if err := tmpl.Execute(w, organization); err != nil {
		log.Panic(err)
	}
}

func sendEvent(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	eventType := r.FormValue("event")

	err := auditlogs.CreateEvent(context.Background(), auditlogs.AuditLogEventOpts{
		Organization: session.Values["org_id"].(string),
		Event: auditlogs.Event{
			Action:     "user." + eventType,
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

	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/send-events", http.StatusSeeOther)
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "session-name")

	if val, ok := session.Values["org_id"].(string); ok {
		// if val is a string
		switch val {
		case "":
			http.Redirect(w, r, "/static/", http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/send-events", http.StatusSeeOther)
		}
	} else {
		// if val is not a string type
		http.Redirect(w, r, "/static/", http.StatusSeeOther)
	}
}

func exportEvents(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	organization := Organization{session.Values["org_name"].(string), session.Values["org_id"].(string)}
	tmpl := template.Must(template.ParseFiles("./static/export_events.html"))
	if err := tmpl.Execute(w, organization); err != nil {
		log.Panic(err)
	}
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	today := time.Now().Format(time.RFC3339)
	lastMonth := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	eventType := r.FormValue("event")

	if eventType == "generate_csv" {
		auditLogExport, err := auditlogs.CreateExport(context.Background(), auditlogs.CreateExportOpts{
			Organization: session.Values["org_id"].(string),
			RangeStart:   lastMonth,
			RangeEnd:     today,
		})

		if err != nil {
			fmt.Println("There was an error generating the CSV:", err)
		}
		session.Values["export_id"] = auditLogExport.Id

		if err := session.Save(r, w); err != nil {
			log.Panic("problem saving cookie", err)
		}

	}

	if eventType == "access_csv" {
		auditLogExport, err := auditlogs.GetExport(context.Background(), auditlogs.GetExportOpts{
			ExportId: session.Values["export_id"].(string),
		})
		http.Redirect(w, r, auditLogExport.Url, http.StatusSeeOther)

		if err != nil {
			fmt.Println("There was an error accessing the CSV:", err)
		}
	}
	http.Redirect(w, r, "/export-events", http.StatusSeeOther)

}

func main() {
	auditlogs.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	router.HandleFunc("/", sessionHandler)

	router.HandleFunc("/get-org", getOrg)
	router.HandleFunc("/send-events", sendEvents)
	router.HandleFunc("/send-event", sendEvent)
	router.HandleFunc("/export-events", exportEvents)
	router.HandleFunc("/get-events", getEvents)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	if err := http.ListenAndServe(":8001", router); err != nil {
		log.Panic(err)
	}
}
