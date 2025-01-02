package userdelete

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Clerk webhook for user deletion")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %s", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	clerkSignature := r.Header.Get("Svix-Signature")
	if clerkSignature == "" {
		http.Error(w, "Missing Svix-Signature header", http.StatusUnauthorized)
		return
	}

	clerkSigningSecret := os.Getenv("DELETE_SIGNING_SECRET")
	if !validateClerkSignature(body, clerkSignature, clerkSigningSecret) {
		http.Error(w, "Invalid Svix-Signature", http.StatusUnauthorized)
		return
	}

	log.Printf("Raw webhook payload: %s", string(body))

	var payload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error unmarshalling JSON: %s", err)
		return
	}
	userID := payload.Data.ID
	log.Printf("Received user ID for deletion: %s", userID)
	if err := DeleteUserFromHasura(userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete user from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error deleting user from Hasura: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "success"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	w.Write(jsonResp)
	log.Printf("User with ID %s successfully deleted from Hasura", userID)
}
func DeleteUserFromHasura(userID string) error {
	query := `
		mutation DeleteUser($id: String!){
			delete_users(where: {id: {_eq: $id}}) {
				affected_rows
			}
		}
	`
	variables := map[string]interface{}{
		"id": userID,
	}
	requestBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal Hasura query: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("hasura responded with status: %s, body: %s", resp.Status, string(body))
	}
	var responseBody struct {
		Data   interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}
	if len(responseBody.Errors) > 0 {
		return fmt.Errorf("hasura errors: %v", responseBody.Errors)
	}
	log.Printf("User with ID %s successfully deleted from Hasura", userID)
	return nil
}

func validateClerkSignature(body []byte, signature, secret string) bool {
	parts := strings.SplitN(signature, ",", 2)
	if len(parts) != 2 {
		log.Println("Invalid signature format")
		return false
	}

	actualSignature := parts[1]
	log.Printf("Extracted Clerk-Signature: %s", actualSignature)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := mac.Sum(nil)

	expectedSignatureBase64 := base64.StdEncoding.EncodeToString(expectedSignature)
	log.Printf("Expected signature (Base64): %s", expectedSignatureBase64)

	if !hmac.Equal([]byte(actualSignature), expectedSignature) {
		log.Printf("Signature mismatch: expected %s, got %s", expectedSignatureBase64, actualSignature)
		return false
	}

	return true
}
