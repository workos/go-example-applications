package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v3/pkg/sso"
)

func init() {
	loadEnvVariables()
}

func loadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Assign the environment variables to the `conf` struct fields
	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.ClientID, "client-id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS client id.")
	flag.StringVar(&conf.RedirectURI, "redirect-uri", os.Getenv("WORKOS_REDIRECT_URI"), "The redirect uri.")
	flag.StringVar(&conf.Connection, "connection", os.Getenv("WORKOS_CONNECTION"), "Use the Connection ID associated with your SSO Connection.")
	flag.StringVar(&conf.Provider, "provider", "", "The OAuth provider used for the SSO connection.")
	flag.Parse()

	log.Printf("launching sso demo: addr=%s client_id=%s redirect_uri=%s connection=%s provider=%s api_key=%s",
		conf.Addr, conf.ClientID, conf.RedirectURI, conf.Connection, conf.Provider, redact(conf.APIKey))

	sso.Configure(conf.APIKey, conf.ClientID)
}

var (
	key    = []byte("super-secret-key")
	store  = sessions.NewCookieStore(key)
	router = http.NewServeMux()
)

type Profile struct {
	FirstName  string
	LastName   string
	RawProfile string
}

var conf struct {
	Addr        string
	APIKey      string
	ClientID    string
	RedirectURI string
	Connection  string
	Provider    string
	APIURL      string
}

func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:8] + "..."
}

func signin(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "cookie-name")
	if err != nil {
		log.Println(err)
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/logged_in", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = false

	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getAuthorizationURL(loginType string) (*url.URL, error) {
	opts := sso.GetAuthorizationURLOpts{
		RedirectURI: conf.RedirectURI,
	}

	if loginType == "saml" {
		opts.Connection = conf.Connection
	} else {
		opts.Provider = sso.ConnectionType(loginType)
	}

	return sso.GetAuthorizationURL(opts)
}

func login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	loginType := r.Form.Get("login_method")

	url, err := getAuthorizationURL(loginType)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}

func callback(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	log.Printf("callback is called with %s", r.URL)

	profile, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOpts{
		Code: r.URL.Query().Get("code"),
	})
	if err != nil {
		log.Printf("get profile failed: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = true
	session.Values["first_name"] = profile.Profile.FirstName
	session.Values["last_name"] = profile.Profile.LastName

	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	raw, _ := json.MarshalIndent(profile, "", "    ")
	thisProfile := Profile{
		FirstName:  profile.Profile.FirstName,
		LastName:   profile.Profile.LastName,
		RawProfile: string(raw),
	}

	if err := tmpl.Execute(w, thisProfile); err != nil {
		log.Panic(err)
	}
}

func loggedin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	session, _ := store.Get(r, "cookie-name")

	thisProfile := Profile{
		FirstName: session.Values["first_name"].(string),
		LastName:  session.Values["last_name"].(string),
	}

	if err := tmpl.Execute(w, thisProfile); err != nil {
		log.Panic(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router.HandleFunc("/login", login)
	router.HandleFunc("/callback", callback)
	router.HandleFunc("/logged_in", loggedin)
	router.HandleFunc("/", signin)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/signin/", http.StripPrefix("/signin/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/logout", logout)

	if err := http.ListenAndServe(conf.Addr, router); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
}
