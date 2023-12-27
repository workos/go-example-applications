package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v3/pkg/organizations"
	"github.com/workos/workos-go/v3/pkg/portal"
)

func ProvisionEnterprise(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}
	organizationDomains := []string{r.FormValue("domain")}
	organizationName := r.FormValue("org")

	organization, err := organizations.CreateOrganization(context.Background(), organizations.CreateOrganizationOpts{
		Name:    organizationName,
		Domains: organizationDomains,
	})

	if err != nil {
		fmt.Println("There was an error creating this organization.")
	}
	//handle logged in
	tmpl := template.Must(template.ParseFiles("./static/admin_logged_in.html"))
	if err := tmpl.Execute(w, organization); err != nil {
		log.Panic(err)
	}
}

func HandlePortal(w http.ResponseWriter, r *http.Request) {
	organizationId := r.URL.Query().Get("id")
	intent := r.URL.Query().Get("intent")

	var linkIntent portal.GenerateLinkIntent
	switch intent {
	case "SSO":
		linkIntent = portal.SSO
	case "Dsync":
		linkIntent = portal.DSync
	case "AuditLogs":
		linkIntent = portal.AuditLogs
	case "LogStreams":
		linkIntent = portal.LogStreams
	default:
		log.Printf("Invalid intent: %s", intent)
		http.Error(w, "Invalid intent", http.StatusBadRequest)
		return
	}

	link, err := portal.GenerateLink(context.Background(), portal.GenerateLinkOpts{
		Organization: organizationId,
		Intent:       linkIntent,
	})
	if err != nil {
		log.Printf("get redirect failed: %s", err)
	}
	http.Redirect(w, r, link, http.StatusFound)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		Addr    string
		Domains string
		APIKey  string
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")

	log.Printf("launching admin portal demo with configuration: %+v", conf)

	organizations.SetAPIKey(conf.APIKey)
	portal.SetAPIKey(conf.APIKey)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/provision-enterprise", ProvisionEnterprise)
	http.HandleFunc("/admin-portal", HandlePortal)

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
