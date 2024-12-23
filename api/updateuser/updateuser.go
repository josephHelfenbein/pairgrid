package updateuser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type UpdateUserRequest struct {
	ID         string   `json:"id"`
	Bio        string   `json:"bio"`
	Language   []string `json:"language"`
	Specialty  []string `json:"specialty"`
	Interests  []string `json:"interests"`
	Occupation string   `json:"occupation"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to update user in Hasura")
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error decoding JSON payload: %s", err)
		return
	}
	if err := UpdateUserInHasura(updateReq); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user in Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error updating user in Hasura: %s", err)
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
	log.Printf("User with ID %s successfully updated in Hasura", updateReq.ID)
}
func UpdateUserInHasura(req UpdateUserRequest) error {
	query := `
		mutation UpdateUser($id: String!, $bio: String, $language: [String!], $specialty: String, $interests: [String!], $occupation: String) {
			update_users_by_pk(
				pk_columns: {id: $id},
				_set: {bio: $bio, language: $language, specialty: $specialty, interests: $interests, occupation: $occupation}
				){
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
		"id":         req.ID,
		"bio":        req.Bio,
		"language":   req.Language,
		"specialty":  req.Specialty,
		"interests":  req.Interests,
		"occupation": req.Occupation,
	}
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON request body: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	reqBody, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	reqBody.Header.Set("Content-Type", "application/json")
	reqBody.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(reqBody)
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
