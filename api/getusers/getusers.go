package getusers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type User struct {
	Name       string   `json:"name"`
	Bio        string   `json:"bio"`
	Language   []string `json:"language"`
	Specialty  string   `json:"specialty"`
	Interests  []string `json:"interests"`
	Occupation string   `json:"occupation"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get user from Hasura")
	users, err := GetUsersFromHasura()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting user from Hasura: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	log.Printf("Users successfully retrieved from Hasura")
}
func GetUsersFromHasura() ([]User, error) {
	query := `
		query GetUsers {
			users(limit: 10) {
				id
				name
				bio
				language
				specialty
				interests
				occupation
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request body: %w", err)
	}
	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var responseBody struct {
		Data struct {
			Users []User `json:"users"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	log.Printf("Hasura response: %+v", responseBody.Data.Users)
	return responseBody.Data.Users, nil
}
