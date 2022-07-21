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
	"github.com/workos/workos-go/pkg/organizations"
	"github.com/workos/workos-go/pkg/portal"
)

func main() {
	var conf struct {
		Addr    string
		Domains string
	}

	type Profile struct {
		First_name  string
		Last_name   string
		Raw_profile string
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	log.Printf("launching admin portal demo with configuration: %+v", conf)
	apiKey := os.Getenv("WORKOS_API_KEY")

	organizations.SetAPIKey(apiKey)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/provision-enterprise", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		organizationDomains := []string{r.FormValue("domain")}
		organizationName := r.FormValue("org")

		organization, err := organizations.CreateOrganization(context.Background(), organizations.CreateOrganizationOpts{
			Name:    organizationName,
			Domains: organizationDomains,
		})

		if err != nil {
			fmt.Println("There's an error")
		}

		//handle logged in
		tmpl := template.Must(template.ParseFiles("./static/admin_logged_in.html"))
		raw_profile := "profile"
		this_profile := Profile{"first", "last", raw_profile}
		tmpl.Execute(w, this_profile)

		http.HandleFunc("/dsync-admin-portal", func(w http.ResponseWriter, r *http.Request) {
			portal.SetAPIKey(apiKey)
			organizationId := organization.ID
			// Generate an SSO Adnim Portal Link using the Organization ID from above.
			link, err := portal.GenerateLink(context.Background(), portal.GenerateLinkOpts{
				Organization: organizationId,
				Intent:       "dsync",
			})
			if err != nil {
				log.Printf("get redirect failed: %s", err)
			}
			http.Redirect(w, r, link, 302)
		})

		http.HandleFunc("/sso-admin-portal", func(w http.ResponseWriter, r *http.Request) {
			portal.SetAPIKey(apiKey)
			organizationId := organization.ID
			// Generate an SSO Adnim Portal Link using the Organization ID from above.
			link, err := portal.GenerateLink(context.Background(), portal.GenerateLinkOpts{
				Organization: organizationId,
				Intent:       "sso",
			})
			if err != nil {
				log.Printf("get redirect failed: %s", err)
			}
			http.Redirect(w, r, link, 302)
		})

	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
