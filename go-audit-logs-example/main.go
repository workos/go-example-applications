package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/auditlogs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		Addr     string
		APIKey   string
		ClientID string
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.APIKey, "client_id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS Client ID.")

	auditlogs.SetAPIKey(conf.APIKey)
	//Action title: "user.signed_in" | Target type: "team"
	//Action title: "user.logged_out" | Target type: "team"
	//Action title: "user.organization_set" | Target type: "team"
	//Action title: "user.organization_deleted" | Target type: "team"
	//Action title: "user.connection_deleted" | Target type: "team"

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}
}
