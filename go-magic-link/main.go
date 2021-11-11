package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/workos-inc/workos-go/pkg/sso"
)

func main() {
	address := ":8000"
	apiKey := os.Getenv("WORKOS_API_KEY")
	clientID := os.Getenv("WORKOS_CLIENT_ID")

	sso.Configure(apiKey, clientID)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		profileAndToken, err := sso.GetProfileAndToken(context.Background(), sso.GetProfileAndTokenOptions{
			Code: r.URL.Query().Get("code"),
		})

		if err != nil {
			// Handle the error ...
		}

		// Use the information in `profile` for further business logic.
		profile := profileAndToken.Profile
		fmt.Println(profile)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Panic(err)
	}
}
