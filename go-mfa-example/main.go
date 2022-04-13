package main

import (
	"context"
	"flag"
	"html/template"
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

	type Response struct {
		ID   string `json:"id"`
		Type string `json:"type"`
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

		this_response := Response{enroll.Type, enroll.ID}
		tmpl := template.Must(template.ParseFiles("./static/enroll_factor.html"))
		tmpl.Execute(w, this_response)

	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
