package getuserinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type GetUserRequest struct {
	Email string `json:"email"`
}

type User struct {
	ID         string   `json:"id"`
	Bio        string   `json:"bio"`
	Language   []string `json:"language"`
	Specialty  string   `json:"specialty"`
	Interests  []string `json:"interests"`
	Occupation string   `json:"occupation"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get user from Hasura")
	var getUserReq GetUserRequest
	if err := json.NewDecoder(r.Body).Decode(&getUserReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error decoding JSON payload: %s", err)
		return
	}
	if getUserReq.Email == "" {
		http.Error(w, "Missing user email", http.StatusBadRequest)
		log.Println("Missing user email")
		return
	}
	user, err := GetUserFromHasura(getUserReq.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting user from Hasura: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	log.Printf("User with email %s successfully retrieved from Hasura", getUserReq.Email)
}
func GetUserFromHasura(userEmail string) (*User, error) {
	query := `
		query GetUser($email: String!) {
			users(where: {email: {_eq: $email}}) {
				id
				bio
				language
				specialty
				interests
				occupation
			}
		}
	`
	variables := map[string]interface{}{
		"email": userEmail,
	}
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body: %w", err)
	}
	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	reqBody, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqBody.Header.Set("Content-Type", "application/json")
	reqBody.Header.Set("x-hasura-admin-secret", hasuraSecret)
	client := &http.Client{}
	resp, err := client.Do(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hasura responded with status: %s", resp.Status)
	}
	var responseBody struct {
		Data struct {
			Users []*User `json:"users"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	if len(responseBody.Data.Users) == 0 {
		return nil, fmt.Errorf("user with email %s not found", userEmail)
	}
	log.Printf("Hasura response: %+v", responseBody.Data.Users[0])
	return responseBody.Data.Users[0], nil
}
