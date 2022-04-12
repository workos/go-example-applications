package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/mfa"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var conf struct {
		APIKey string
	}

	type Users struct {
		Users string
	}

	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.Parse()

	mfa.SetAPIKey("sk_test_a2V5XzAxRkFLNFJLSzVBUTBSRTNDVFFURFhEQks1LHZwWERmTlUxZUNlUUM0Y1Nrb09LNndpQ3M")

	enroll, err := mfa.EnrollFactor(context.Background(), mfa.GetEnrollOpts{
		Type:       "totp",
		TotpIssuer: "WorkOS",
		TotpUser:   "some_user",
	})
	if err != nil {
		log.Printf("enroll factor failed: %s", err)
		return
	}

	b, err := json.MarshalIndent(enroll, "", "    ")
	if err != nil {
		log.Printf("error: %s", err)
	}

	// Stringify the returned users
	stringB := string(b)
	data := Users{stringB}

	// Render the template with the users as data
	fmt.Println(data)

}
