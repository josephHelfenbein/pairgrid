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
	"strconv"
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

	svixTimestamp := r.Header.Get("Svix-Timestamp")
	if svixTimestamp == "" {
		http.Error(w, "Missing Svix-Timestamp header", http.StatusUnauthorized)
		return
	}

	clerkSigningSecret := os.Getenv("DELETE_SIGNING_SECRET")
	if !validateClerkSignature(body, clerkSignature, clerkSigningSecret, r) {
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

func validateClerkSignature(body []byte, signature, secret string, r *http.Request) bool {
	if secret == "" {
		log.Println("Signing secret is empty")
		return false
	}

	secretBytes, err := base64.StdEncoding.DecodeString(strings.Split(secret, "_")[1])
	if err != nil {
		log.Printf("Failed to base64 decode the secret: %v", err)
		return false
	}

	log.Printf("Received signature header: %s", signature)
	log.Printf("Received body: %s", string(body))

	svixTimestamp := r.Header.Get("Svix-Timestamp")
	if svixTimestamp == "" {
		log.Println("Svix-Timestamp header is missing")
		return false
	}
	log.Printf("Received timestamp: %s", svixTimestamp)

	svixTimestampInt, err := strconv.ParseInt(svixTimestamp, 10, 64)
	if err != nil {
		log.Printf("Error parsing timestamp: %v", err)
		return false
	}
	if svixTimestampInt < 10000000000 {
		svixTimestampInt *= 1000
	}

	message := fmt.Sprintf("%d.%s", svixTimestampInt, string(body))
	log.Printf("Constructed message to sign: %s", message)

	signatureParts := strings.SplitN(signature, ",", 2)
	if len(signatureParts) != 2 || signatureParts[0] != "v1" {
		log.Println("Invalid signature format or version")
		return false
	}

	providedSignature := signatureParts[1]
	log.Printf("Extracted signature: %s", providedSignature)

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(message))
	computedMAC := mac.Sum(nil)

	decodedSignature, err := base64.StdEncoding.DecodeString(providedSignature)
	if err != nil {
		log.Printf("Failed to decode signature: %v", err)
		return false
	}

	isValid := hmac.Equal(decodedSignature, computedMAC)
	log.Printf("Signature validation result: %v", isValid)

	return isValid
}
