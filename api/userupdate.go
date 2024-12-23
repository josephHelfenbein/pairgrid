package handler

import (
	"api/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Clerk webhook request")

	if err := utils.CreateUserInHasura(user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user into Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error inserting user into Hasura: %s", err)
		return
	}

	var user utils.ClerkUser
	if err := json.Unmarshal(body, &user); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error unmarshalling JSON: %s", err)
		return
	}

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
