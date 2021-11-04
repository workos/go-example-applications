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
	"github.com/workos-inc/workos-go/pkg/sso"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		Addr        string
		APIKey      string
		ClientID    string
		RedirectURI string
		Domain      string
		Provider    string
	}

	type Profile struct {
		First_name  string
		Last_name   string
		Raw_profile string
	}

	flag.StringVar(&conf.Addr, "addr", ":3042", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.ClientID, "client-id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS client id.")
	flag.StringVar(&conf.RedirectURI, "redirect-uri", os.Getenv("WORKOS_REDIRECT_URI"), "The redirect uri.")
	flag.StringVar(&conf.Domain, "domain", os.Getenv("WORKOS_DOMAIN"), "The domain used to register a WorkOS SSO connection.")
	flag.StringVar(&conf.Provider, "provider", "", "The OAuth provider used for the SSO connection.")
	flag.Parse()

	log.Printf("launching sso demo with configuration: %+v", conf)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Configure the WorkOS SSO SDK:
	sso.Configure(conf.APIKey, conf.ClientID)

	// Handle login
	http.Handle("/login", sso.Login(sso.GetAuthorizationURLOptions{
		//Instead of domain, you can now use connection ID to associate a user to the appropriate connection.
		Domain:      conf.Domain,
		RedirectURI: conf.RedirectURI,
	}))

	// Handle login redirect:
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("callback is called with %s", r.URL)

		// Retrieving user profile:
		profile, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOptions{
			Code: r.URL.Query().Get("code"),
		})
		if err != nil {
			log.Printf("get profile failed: %s", err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// Display user profile:
		b, err := json.MarshalIndent(profile, "", "    ")
		if err != nil {
			log.Printf("encoding profile failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		// define variable to hold data
		var data map[string]interface{}

		// decode the json and set pointer to the data variable
		if err := json.Unmarshal(b, &data); err != nil {
			panic(err)
		}

		// Unnest the profile
		first_name := data["profile"].(map[string]interface{})["first_name"]
		last_name := data["profile"].(map[string]interface{})["last_name"]

		// Convert to strings
		first_name_string := first_name.(string)
		last_name_string := last_name.(string)
		raw_profile := string(b)

		// Create instance of Profile struct including profile values
		this_profile := Profile{first_name_string, last_name_string, raw_profile}

		// Render the template
		tmpl.Execute(w, this_profile)
	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
