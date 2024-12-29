package getmessages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Message struct {
	ID               string `json:"id"`
	SenderID         string `json:"sender_id"`
	RecipientID      string `json:"recipient_id"`
	EncryptedContent string `json:"encrypted_content"`
	CreatedAt        string `json:"created_at"`
	Key              string `json:"key"`
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get messages from Hasura")

	query := r.URL.Query()
	senderID := query.Get("user_id")
	recipientID := query.Get("friend_id")

	if senderID == "" || recipientID == "" {
		http.Error(w, "Missing user_id or friend_id query parameter", http.StatusBadRequest)
		return
	}

	messages, err := GetMessages(senderID, recipientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get messages from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting messages from Hasura: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	log.Printf("Messages successfully retrieved from Hasura")
}

func GetMessages(senderID, recipientID string) ([]Message, error) {
	query := `
		query GetMessages($senderID: String!, $recipientID: String!) {
			messages(
				where: {
					_or: [
						{ sender_id: { _eq: $senderID }, recipient_id: { _eq: $recipientID } },
						{ sender_id: { _eq: $recipientID }, recipient_id: { _eq: $senderID } }
					]
				},
				order_by: { created_at: asc }
			) {
				id
				sender_id
				recipient_id
				encrypted_content
				created_at
				key
			}
		}
	`

	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"senderID":    senderID,
			"recipientID": recipientID,
		},
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

	client := &http.Client{Timeout: 10 * time.Second}
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
			Messages []Message `json:"messages"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return responseBody.Data.Messages, nil
}
