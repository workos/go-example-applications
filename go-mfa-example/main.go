package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"os"
	"fmt"
	"encoding/json"
	"encoding/gob"
	"context"
	"html/template"
	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v2/pkg/mfa"
	"github.com/gorilla/sessions"
)

var router = http.NewServeMux()
var key = []byte("super-secret-key")
var store = sessions.NewCookieStore(key)

var conf struct {
	APIKey string
	Addr   string
}

type Cookie struct {
	Type        string
	ID			string
	Phone       string
	CreatedAt   string
	UpdatedAt   string
}

type Factor struct {
    Type  string
    ID string
}


func displayFactors(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/index.html"))
	session, err := store.Get(r, "cookie-name")
	
	if err != nil {
		fmt.Printf("Error getting session: %v\n", err)
	}

	data := session.Values["factors"]

	// Render the template with the directories
	tmpl.Execute(w, data)
}

func enrollFactor(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/enroll_factors.html"))
	data := "test"
	// Render the template with the directories
	tmpl.Execute(w, data)
}

type EnrollRequest struct {
    Type        mfa.FactorType `json:"type"`
    TOTPIssuer  string         `json:"issuer"`
    TOTPUser    string         `json:"user"`
	PhoneNumber 		string			`json:"phone_number"`
}

func enroll(enrollOpts mfa.EnrollFactorOpts) (mfa.Factor, error) {
    enroll, err := mfa.EnrollFactor(context.Background(), enrollOpts)
    return enroll, err
}

func enrollHandler(w http.ResponseWriter, r *http.Request) {
    var req EnrollRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
       fmt.Println(err)
    }

    enrollOpts := mfa.EnrollFactorOpts{
        Type:        req.Type,
        TOTPIssuer:  req.TOTPIssuer,
        TOTPUser:    req.TOTPUser,
        PhoneNumber: req.PhoneNumber,
    }


    enrollResponse, err := enroll(enrollOpts)
    if err != nil {
        fmt.Println(err)
    }

	cookie := Cookie{
		Type:        string(req.Type),
		ID:          enrollResponse.ID,
		Phone:       req.PhoneNumber,
		CreatedAt:   enrollResponse.CreatedAt,
		UpdatedAt:   enrollResponse.UpdatedAt,
	}

	// Get the session
	session, err := store.Get(r, "cookie-name")
	if err != nil {
		fmt.Println(err)
	}

	factorsInterface := session.Values["factors"]

	if factorsInterface == nil {
		factorsInterface = []Cookie{}
	}

factors, _ := factorsInterface.([]Cookie)

// Append the new cookie to the factors array
factors = append(factors, cookie)

// Set the updated factors slice back into the session
session.Values["factors"] = factors

	err = session.Save(r, w)
	if err != nil {
		fmt.Println(err)
	}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(enrollResponse)
}

func factorDetail(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/factor_detail.html"))
	id := r.URL.Query().Get("id")

	session, err := store.Get(r, "cookie-name")
	
	if err != nil {
		fmt.Printf("Error getting session: %v\n", err)
	}

	// Get the factors from the session
	factors := session.Values["factors"]

	factorsSlice, ok := factors.([]Cookie)
	if !ok {
		fmt.Println("Problem with Factors Slice")
	}

	// Loop through the sessions to find the object that matches the ID
	for _, s := range factorsSlice {
		if s.ID == id {
			// The object that matches the ID is found
			data := s
			tmpl.Execute(w, data)
		}
	}

}

func challengeFactor (w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/challenge_factor.html"))
	id := r.URL.Query().Get("id")
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	smsMessage := r.FormValue("sms_message")

	challenge, err := mfa.ChallengeFactor(context.Background(), mfa.ChallengeFactorOpts{
		FactorID:    id,
		SMSTemplate: smsMessage,
	})
	
	if err != nil {
		fmt.Println("Challenge err: %v\n", err)
	}

	tmpl.Execute(w, challenge.ID)

}

func verifyFactor (w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/challenge_success.html"))
	id := r.URL.Query().Get("id")
	
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	//get code from form
	rawCode := []string{r.FormValue("code-1"), r.FormValue("code-2"), r.FormValue("code-3"), r.FormValue("code-4"), r.FormValue("code-5"), r.FormValue("code-6")}
	code := strings.Join(rawCode, "")

	verify, err := mfa.VerifyChallenge(context.Background(), mfa.VerifyChallengeOpts{
		ChallengeID: id,
		Code:        code,
	})

	if err != nil {
		fmt.Println(err)
		//execute file with error
		tmpl.Execute(w, err)
	}

	tmpl.Execute(w, verify)
}

func clearSession (w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["factors"] = ""
	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	gob.Register([]Cookie{})
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.Parse()

	log.Printf("launching mfa demo with configuration: %+v", conf)

	mfa.SetAPIKey(conf.APIKey)

	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/", displayFactors)

	router.HandleFunc("/enroll-factors", enrollFactor)
	router.HandleFunc("/enroll-factor", enrollHandler)
	router.HandleFunc("/factor-detail", factorDetail)
	router.HandleFunc("/challenge-factor", challengeFactor)
	router.HandleFunc("/verify-factor", verifyFactor)
	router.HandleFunc("/clear-session", clearSession)


	

	if err := http.ListenAndServe(conf.Addr, router); err != nil {
		log.Panic(err)
	}
}
