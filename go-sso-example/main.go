package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/sso"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key    = []byte("super-secret-key")
	store  = sessions.NewCookieStore(key)
	router = http.NewServeMux()
)

type Profile struct {
	First_name  string
	Last_name   string
	Raw_profile string
}

func signin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Redirect to Profile if Logged in
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
		Connection  string
		Provider    string
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.ClientID, "client-id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS client id.")
	flag.StringVar(&conf.RedirectURI, "redirect-uri", os.Getenv("WORKOS_REDIRECT_URI"), "The redirect uri.")
	flag.StringVar(&conf.Connection, "connection", os.Getenv("WORKOS_CONNECTION"), "Use the Connection ID associated to your SSO Connection..")
	flag.StringVar(&conf.Provider, "provider", "", "The OAuth provider used for the SSO connection.")
	flag.Parse()

	log.Printf("launching sso demo with configuration: %+v", conf)

	// Configure the WorkOS SSO SDK:
	sso.Configure(conf.APIKey, conf.ClientID)

	// Handle login
	router.Handle("/login", sso.Login(sso.GetAuthorizationURLOptions{
		Connection:  conf.Connection,
		RedirectURI: conf.RedirectURI,
	}))

	// Handle login redirect:
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	router.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {

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
		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = true
		session.Save(r, w)

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

	// Handle routes:
	router.HandleFunc("/", signin)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/logout", logout)
	http.ListenAndServe(conf.Addr, router)
}
