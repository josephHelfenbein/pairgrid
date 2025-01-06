package getmessages

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
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

type Message struct {
	ID               string `json:"id"`
	SenderID         string `json:"sender_id"`
	RecipientID      string `json:"recipient_id"`
	EncryptedContent string `json:"encrypted_content"`
	CreatedAt        string `json:"created_at"`
	Key              string `json:"key"`
}

func DecryptMessage(encryptedContent, ivHex string, key []byte) (string, error) {
	cipherText, err := hex.DecodeString(encryptedContent)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}
	if len(iv) != aes.BlockSize {
		return "", fmt.Errorf("invalid IV length: expected %d bytes, got %d", aes.BlockSize, len(iv))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	plainText := make([]byte, len(cipherText))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
func GenerateEncryptionKey(userID, serverSecret string) []byte {
	hash := sha256.Sum256([]byte(userID + serverSecret))
	return hash[:]
}
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get messages from Hasura")
	clerk.SetKey(os.Getenv("NUXT_CLERK_SECRET_KEY"))
	sessionToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	log.Printf("Found session token %s", sessionToken)
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
	log.Printf("Found user %s", usr.ID)

	query := r.URL.Query()
	senderID := query.Get("user_id")
	recipientID := query.Get("friend_id")

	if senderID == "" || recipientID == "" {
		http.Error(w, "Missing user_id or friend_id query parameter", http.StatusBadRequest)
		return
	}
	if senderID != usr.ID {
		http.Error(w, "JWT subject does not match request ID", http.StatusForbidden)
		log.Printf("JWT subject (%s) does not match request ID (%s)", usr.ID, senderID)
		return
	}
	messages, err := GetMessages(senderID, recipientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get messages from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting messages from Hasura: %s", err)
		return
	}
	err = CheckAndUpdateNotifications(recipientID, senderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update notifications: %s", err), http.StatusInternalServerError)
		log.Printf("Error updating notifications: %s", err)
	}
	serverSecret := os.Getenv("ENCRYPTION_KEY")

	for i, message := range messages {
		key := GenerateEncryptionKey(message.SenderID, serverSecret)
		decrypted, err := DecryptMessage(message.EncryptedContent, message.Key, key)
		if err != nil {
			log.Printf("Failed to decrypt message ID %s: %s", message.ID, err)
			continue
		}
		messages[i].EncryptedContent = decrypted
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

func CheckAndUpdateNotifications(senderID, recipientID string) error {
	fetchQuery := `
		query GetNotifications($recipientID: String!) {
			notifications(where: { user: { _eq: $recipientID } }) {
				from_users
			}
		}
	`

	fetchRequestBody := map[string]interface{}{
		"query": fetchQuery,
		"variables": map[string]interface{}{
			"recipientID": recipientID,
		},
	}

	fetchJSONBody, err := json.Marshal(fetchRequestBody)
	if err != nil {
		return fmt.Errorf("failed to create request body for fetching notifications: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	fetchReq, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(fetchJSONBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	fetchReq.Header.Set("Content-Type", "application/json")
	fetchReq.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	fetchResp, err := client.Do(fetchReq)
	if err != nil {
		return fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer fetchResp.Body.Close()

	if fetchResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", fetchResp.StatusCode)
	}

	var fetchResponseBody struct {
		Data struct {
			Notifications []struct {
				FromUsers []string `json:"from_users"`
			} `json:"notifications"`
		} `json:"data"`
	}
	if err := json.NewDecoder(fetchResp.Body).Decode(&fetchResponseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	if len(fetchResponseBody.Data.Notifications) == 0 {
		return fmt.Errorf("no notifications found for recipientID: %s", recipientID)
	}

	notification := fetchResponseBody.Data.Notifications[0]
	fromUsers := notification.FromUsers
	updatedFromUsers := []string{}
	for _, user := range fromUsers {
		if user != senderID {
			updatedFromUsers = append(updatedFromUsers, user)
		}
	}

	updateMutation := `
		mutation UpdateNotifications($recipientID: String!, $fromUsers: [String!]) {
			update_notifications(
				where: { user: { _eq: $recipientID } },
				_set: { from_users: $fromUsers }
			) {
				affected_rows
			}
		}
	`

	updateRequestBody := map[string]interface{}{
		"query": updateMutation,
		"variables": map[string]interface{}{
			"recipientID": recipientID,
			"fromUsers":   updatedFromUsers,
		},
	}
	updateJSONBody, err := json.Marshal(updateRequestBody)
	if err != nil {
		return fmt.Errorf("failed to create request body for notification update: %w", err)
	}

	updateReq, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(updateJSONBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("x-hasura-admin-secret", hasuraSecret)

	updateResp, err := client.Do(updateReq)
	if err != nil {
		return fmt.Errorf("failed to send update request to Hasura: %w", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", updateResp.StatusCode)
	}

	var updateResponseBody struct {
		Data struct {
			UpdateNotifications struct {
				AffectedRows int `json:"affected_rows"`
			} `json:"update_notifications"`
		} `json:"data"`
	}
	if err := json.NewDecoder(updateResp.Body).Decode(&updateResponseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	log.Printf("Updated notifications: %d rows affected", updateResponseBody.Data.UpdateNotifications.AffectedRows)
	return nil
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
