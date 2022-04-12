package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
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
		Addr   string
	}

	type Users struct {
		Users string
	}

	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Addr, "addr", ":8002", "The server addr.")
	flag.Parse()

	mfa.SetAPIKey(conf.APIKey)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/enroll-factor", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		enrolltype := r.FormValue("type")
		totpissuer := r.FormValue("totp_issuer")
		totpuser := r.FormValue("totp_user")
		number := r.FormValue("phone_number")

		fmt.Println(enrolltype)

		enroll, err := mfa.EnrollFactor(context.Background(), mfa.GetEnrollOpts{
			Type:        enrolltype,
			TotpIssuer:  totpissuer,
			TotpUser:    totpuser,
			PhoneNumber: number,
		})
		if err != nil {
			log.Printf("enroll factor failed: %s", err)
			return
		}

		b, err := json.MarshalIndent(enroll, "", "    ")
		if err != nil {
			log.Printf("error: %s", err)
		}

		stringB := string(b)
		data := Users{stringB}

		fmt.Println(data)
	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
