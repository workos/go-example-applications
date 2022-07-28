package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/workos/workos-go/pkg/mfa"
)

var sessions = map[string]session{}
var router = http.NewServeMux()

type session struct {
	enrollID    string
	challengeID string
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
	Totp        interface{}
}

type VerifyResponse struct {
	Valid     bool `json:"valid"`
	Challenge interface{}
}

func enrollHandler(w http.ResponseWriter, r *http.Request) {
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
	qrCode := fmt.Sprint(enroll.Totp["qr_code"])

	sessions["your_token"] = session{
		enrollID: enroll.ID,
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: "your_token",
	})

	this_response := Response{enroll.ID, enroll.Type, enroll.CreatedAt, enroll.UpdatedAt, enroll.Object, SmsDetails["phone_number"], template.URL(qrCode)}
	tmpl := template.Must(template.ParseFiles("./static/factor_detail.html"))
	tmpl.Execute(w, this_response)
}

func challengeFactor(w http.ResponseWriter, r *http.Request) {

	smstemplate := r.FormValue("sms_message")

	challenge, err := mfa.ChallengeFactor(context.Background(), mfa.ChallengeOpts{
		AuthenticationFactorID: sessions["your_token"].enrollID,
		SMSTemplate:            smstemplate,
	})

	if err != nil {
		log.Printf("challengefailed: %s", err)
		return
	}

	sessions["your_token"] = session{
		challengeID: challenge.ID,
	}

	tmpl := template.Must(template.ParseFiles("./static/challenge_factor.html"))
	tmpl.Execute(w, "")
}

func verifyFactor(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	verify, err := mfa.VerifyFactor(context.Background(), mfa.VerifyOpts{
		AuthenticationChallengeID: sessions["your_token"].challengeID,
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
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Addr, "addr", ":8002", "The server addr.")
	flag.Parse()

	mfa.SetAPIKey(conf.APIKey)

	router.Handle("/", http.FileServer(http.Dir("./static")))
	router.HandleFunc("/enroll-factor", enrollHandler)
	router.HandleFunc("/challenge-factor", challengeFactor)
	router.HandleFunc("/verify-factor", verifyFactor)

	http.ListenAndServe(conf.Addr, router)
}
