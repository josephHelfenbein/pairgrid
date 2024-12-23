package handler

import (
	"api/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Clerk webhook request")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %s", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Raw webhook payload: %s", string(body))
	var rawPayload struct {
		Data struct {
			ID               string                  `json:"id"`
			FirstName        string                  `json:"first_name"`
			LastName         string                  `json:"last_name"`
			EmailAddresses   []utils.EmailAddress    `json:"email_addresses"`
			ExternalAccounts []utils.ExternalAccount `json:"external_accounts"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &rawPayload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error unmarshalling JSON: %s", err)
		return
	}
	user := utils.ClerkUser{
		ID:               rawPayload.Data.ID,
		FirstName:        rawPayload.Data.FirstName,
		LastName:         rawPayload.Data.LastName,
		EmailAddresses:   rawPayload.Data.EmailAddresses,
		ExternalAccounts: rawPayload.Data.ExternalAccounts,
	}
	log.Printf("Received user: %+v", user)

	if err := utils.CreateUserInHasura(user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user into Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error inserting user into Hasura: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "success"}
	if jsonResp, err := json.Marshal(resp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	} else {
		w.Write(jsonResp)
	}

	log.Printf("User with ID %s successfully created in Hasura", user.ID)
}
