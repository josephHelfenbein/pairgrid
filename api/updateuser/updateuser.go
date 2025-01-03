package updateuser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

type UpdateUserRequest struct {
	ID         string   `json:"id"`
	Bio        string   `json:"bio"`
	Language   []string `json:"language"`
	Specialty  string   `json:"specialty"`
	Interests  []string `json:"interests"`
	Occupation string   `json:"occupation"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to update user in Hasura")
	clerk.SetKey(os.Getenv("NUXT_CLERK_SECRET_KEY"))
	sessionToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
		Token: sessionToken,
	})
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		log.Printf("Session not found")
		return
	}
	usr, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		http.Error(w, "User could not be retrieved from session", http.StatusUnauthorized)
		log.Printf("User could not be retrieved from session")
		return
	}
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error decoding JSON payload: %s", err)
		return
	}
	if usr.ID != updateReq.ID {
		http.Error(w, "JWT subject does not match request ID", http.StatusForbidden)
		log.Printf("JWT subject (%s) does not match request ID (%s)", usr.ID, updateReq.ID)
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
		mutation UpdateUser($id: String!, $bio: String, $language: [String!], $specialty: String, $interests: [String!], $occupation: String, $lastSeen: timestamptz!) {
			update_users_by_pk(
				pk_columns: {id: $id},
				_set: {bio: $bio, language: $language, specialty: $specialty, interests: $interests, occupation: $occupation, last_seen: $lastSeen}
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
		"lastSeen":   time.Now().Format(time.RFC3339Nano),
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
