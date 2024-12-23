package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type ClerkUser struct {
	ID               string            `json:"id"`
	FirstName        string            `json:"first_name"`
	LastName         string            `json:"last_name"`
	EmailAddresses   []EmailAddress    `json:"email_addresses"`
	ExternalAccounts []ExternalAccount `json:"external_accounts"`
	ImageURL         string            `json:"image_url"`
	LastActiveAt     int64             `json:"last_active_at"`
	LastSignInAt     *int64            `json:"last_sign_in_at,omitempty"`
	Locked           bool              `json:"locked"`
	Username         string            `json:"username"`
}

type EmailAddress struct {
	EmailAddress string `json:"email_address"`
	ID           string `json:"id"`
	Verification struct {
		Status string `json:"status"`
	} `json:"verification"`
}

type ExternalAccount struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

func quoteIfNotEmpty(s string) string {
	if s == "" {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, s)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to create user in Hasura")

	var user ClerkUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error unmarshalling JSON: %s", err)
		return
	}

	if err := CreateUserInHasura(user); err != nil {
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

func CreateUserInHasura(user ClerkUser) error {
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	quotedFullName := quoteIfNotEmpty(fullName)
	quotedEmail := quoteIfNotEmpty(user.EmailAddresses[0].EmailAddress)
	query := fmt.Sprintf(`
		mutation {
			insert_users(objects: {id: "%s", name: %s, email: "%s"}) {
				returning {
					id
					name
					email
				}
			}
		}
	`, user.ID, quotedFullName, quotedEmail)

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal Hasura query: %w", err)
	}

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hasura responded with status: %s", resp.Status)
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	log.Printf("Hasura response: %+v", responseBody)
	return nil
}
