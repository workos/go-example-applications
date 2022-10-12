package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/auditlogs"
)

var router = http.NewServeMux()

var conf struct {
	Addr     string
	APIKey   string
	ClientID string
}

func setOrg(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	org := r.FormValue("org")

	fmt.Println(org)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.APIKey, "client_id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS Client ID.")

	auditlogs.SetAPIKey(conf.APIKey)

	http.Handle("/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/set-org", setOrg)
	//Action title: "user.signed_in" | Target type: "team"
	//Action title: "user.logged_out" | Target type: "team"
	//Action title: "user.organization_set" | Target type: "team"
	//Action title: "user.organization_deleted" | Target type: "team"
	//Action title: "user.connection_deleted" | Target type: "team"

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
