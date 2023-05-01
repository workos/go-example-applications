package main

import (
	"flag"
	"log"
	"net/http"
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
	session, _ := store.Get(r, "cookie-name")
	
	fmt.Println(session.Values["factors"])



	// Render the template with the directories
	tmpl.Execute(w, "test")
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

factors, ok := factorsInterface.([]Cookie)
if !ok {
    fmt.Println("Factors not ok")
}

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
	

	if err := http.ListenAndServe(conf.Addr, router); err != nil {
		log.Panic(err)
	}
}
