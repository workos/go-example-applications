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
	"github.com/workos/workos-go/v2/pkg/organizations"
	"github.com/workos/workos-go/v2/pkg/portal"
)

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

	type Profile struct {
		First_name  string
		Last_name   string
		Raw_profile string
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")

	log.Printf("launching admin portal demo with configuration: %+v", conf)

	organizations.SetAPIKey(conf.APIKey)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/provision-enterprise", func(w http.ResponseWriter, r *http.Request) {
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
		raw_profile := "profile"
		this_profile := Profile{"first", "last", raw_profile}
		if err := tmpl.Execute(w, this_profile); err != nil {
			log.Panic(err)
		}

		http.HandleFunc("/dsync-admin-portal", func(w http.ResponseWriter, r *http.Request) {
			portal.SetAPIKey(conf.APIKey)
			organizationId := organization.ID
			// Generate an SSO Adnim Portal Link using the Organization ID from above.
			link, err := portal.GenerateLink(context.Background(), portal.GenerateLinkOpts{
				Organization: organizationId,
				Intent:       "dsync",
			})
			if err != nil {
				log.Printf("get redirect failed: %s", err)
			}
			http.Redirect(w, r, link, http.StatusFound)
		})

		http.HandleFunc("/sso-admin-portal", func(w http.ResponseWriter, r *http.Request) {
			portal.SetAPIKey(conf.APIKey)
			organizationId := organization.ID
			// Generate an SSO Adnim Portal Link using the Organization ID from above.
			link, err := portal.GenerateLink(context.Background(), portal.GenerateLinkOpts{
				Organization: organizationId,
				Intent:       "sso",
			})
			if err != nil {
				log.Printf("get redirect failed: %s", err)
			}
			http.Redirect(w, r, link, http.StatusFound)
		})

	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
