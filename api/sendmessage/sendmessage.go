package sendmessage

import (
	"api/addfriend"
	"api/updateseen"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Message struct {
	SenderID      string `json:"sender_id"`
	ReceiverEmail string `json:"receiver_email"`
	Content       string `json:"content"`
	Key           string `json:"key"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error decoding JSON payload: %s", err)
		return
	}
	if msg.SenderID == "" || msg.ReceiverEmail == "" || msg.Content == "" || msg.Key == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		log.Printf("Missing fields: %+v", msg)
		return
	}
	updateseen.UpdateUserInHasura(msg.SenderID)
	receiverID, err := addfriend.GetFriendIDByEmail(msg.ReceiverEmail)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get receiver ID: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting receiver ID: %s", err)
		return
	}
	if err := InsertMessage(msg.SenderID, receiverID, msg.Content, msg.Key); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert message: %s", err), http.StatusInternalServerError)
		log.Printf("Error inserting message: %s", err)
		return
	}
	response := map[string]string{"status": "success"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.Printf("Message sent from %s to %s", msg.SenderID, msg.ReceiverEmail)
}
func InsertMessage(senderID, retrieverID, content, key string) error {
	query := `
		mutation InsertMessages($senderID: String!, $recipientID: String!, $content: String!, $key: String!) {
			insert_messages(objects: {sender_id: $senderID, recipient_id: $recipientID, encrypted_content: $content, encrypted_key: $key}) {
				affected_rows
			}
		}
	`

	variables := map[string]interface{}{
		"senderID":    senderID,
		"recipientID": retrieverID,
		"content":     content,
		"key":         key,
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

	log.Printf("Sent message in Hasura")
	return nil
}
