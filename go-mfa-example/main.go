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
		ID          string `json:"id"`
		Type        string `json:"type"`
		Environment string `json:"environment_id`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Phone       interface{}
	}

	type VerifyResponse struct {
		Valid     bool `json:"valid"`
		Challenge interface{}
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

		SmsDetails := enroll.Sms

		this_response := Response{enroll.ID, enroll.Type, enroll.EnvironmentID, enroll.CreatedAt, enroll.UpdatedAt, SmsDetails["phone_number"]}
		tmpl := template.Must(template.ParseFiles("./static/enroll_factor.html"))
		tmpl.Execute(w, this_response)

		http.HandleFunc("/factor-detail", func(w http.ResponseWriter, r *http.Request) {
			this_response := Response{enroll.ID, enroll.Type, enroll.CreatedAt, enroll.UpdatedAt, enroll.EnvironmentID, SmsDetails["phone_number"]}
			tmpl := template.Must(template.ParseFiles("./static/factor_detail.html"))
			tmpl.Execute(w, this_response)
		})

		http.HandleFunc("/challenge-factor", func(w http.ResponseWriter, r *http.Request) {
			smstemplate := r.FormValue("sms_message")

			challenge, err := mfa.ChallengeFactor(context.Background(), mfa.ChallengeOpts{
				AuthenticationFactorID: enroll.ID,
				SMSTemplate:            smstemplate,
			})

			if err != nil {
				log.Printf("challengefailed: %s", err)
				return
			}

			tmpl := template.Must(template.ParseFiles("./static/challenge_factor.html"))
			tmpl.Execute(w, "")

			http.HandleFunc("/verify-factor", func(w http.ResponseWriter, r *http.Request) {
				code := r.FormValue("code")

				verify, err := mfa.VerifyFactor(context.Background(), mfa.VerifyOpts{
					AuthenticationChallengeID: challenge.ID,
					Code:                      code,
				})

				if err != nil {
					log.Printf("challengefailed: %s", err)
					return
				}

				valid := verify.(mfa.VerifyResponse).Valid
				challenge := verify.(mfa.VerifyResponse).Challenge
				tmpl := template.Must(template.ParseFiles("./static/challenge_success.html"))
				this_response := VerifyResponse{valid, challenge}
				tmpl.Execute(w, this_response)
			})
		})

	})

	if err := http.ListenAndServe(conf.Addr, nil); err != nil {
		log.Panic(err)
	}

}
