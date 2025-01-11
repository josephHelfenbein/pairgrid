package sendmessage

import (
	"api/addfriend"
	"api/updateseen"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/pusher/pusher-http-go/v5"
)

type Message struct {
	SenderID      string `json:"sender_id"`
	ReceiverEmail string `json:"receiver_email"`
	Content       string `json:"content"`
	Key           string `json:"key"`
}
type MessagePusher struct {
	SenderID         string `json:"sender_id"`
	RecipientID      string `json:"recipient_id"`
	EncryptedContent string `json:"encrypted_content"`
	Key              string `json:"key"`
	CreatedAt        string `json:"created_at"`
}
type VoiceCall struct {
	CallerID string `json:"caller_id"`
	CalleeID string `json:"callee_id"`
	Type     string `json:"type"`
}

func GenerateEncryptionKey(userID, serverSecret string) []byte {
	hash := sha256.Sum256([]byte(userID + serverSecret))
	return hash[:]
}
func EncryptMessage(plainText string, key []byte) (string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", fmt.Errorf("failed to generate IV: %w", err)
	}
	cipherText := make([]byte, len(plainText))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText, []byte(plainText))
	return hex.EncodeToString(cipherText), hex.EncodeToString(iv), nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
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

	var msg Message
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil || msg.SenderID == "" || msg.ReceiverEmail == "" || msg.Content == "" {
		var voicecall VoiceCall
		err = json.NewDecoder(r.Body).Decode(&voicecall)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
			log.Printf("Error decoding JSON payload: %s", err)
			return
		}
		if voicecall.CallerID == "" || voicecall.CalleeID == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			log.Printf("Missing fields: %+v", msg)
			return
		}
		if voicecall.CallerID != usr.ID {
			http.Error(w, "JWT subject does not match request ID", http.StatusForbidden)
			log.Printf("JWT subject (%s) does not match request ID (%s)", usr.ID, voicecall.CallerID)
			return
		}
		BroadcastVoiceCall(voicecall)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "call request sent"})
		return
	}
	if msg.SenderID != usr.ID {
		http.Error(w, "JWT subject does not match request ID", http.StatusForbidden)
		log.Printf("JWT subject (%s) does not match request ID (%s)", usr.ID, msg.SenderID)
		return
	}
	updateseen.UpdateUserInHasura(msg.SenderID)
	receiverID, err := addfriend.GetFriendIDByEmail(msg.ReceiverEmail)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get receiver ID: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting receiver ID: %s", err)
		return
	}
	serverSecret := os.Getenv("ENCRYPTION_KEY")
	encryptionKey := GenerateEncryptionKey(msg.SenderID, serverSecret)
	encryptedContent, iv, err := EncryptMessage(msg.Content, encryptionKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encrypt message: %s", err), http.StatusInternalServerError)
		log.Printf("Error encrypting message: %s", err)
		return
	}
	BroadcastMessage(MessagePusher{
		SenderID:         msg.SenderID,
		RecipientID:      receiverID,
		EncryptedContent: msg.Content,
		CreatedAt:        time.Now().Format(time.RFC3339Nano),
	})
	msg.Content = encryptedContent
	msg.Key = iv
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
	createdAt := time.Now().Format(time.RFC3339Nano)
	query := `
		mutation InsertMessages($senderID: String!, $recipientID: String!, $content: String!, $key: String!, $createdAt: timestamptz!) {
			insert_messages(objects: {sender_id: $senderID, recipient_id: $recipientID, encrypted_content: $content, key: $key, created_at: $createdAt}) {
				affected_rows
			}
		}
	`

	variables := map[string]interface{}{
		"senderID":    senderID,
		"recipientID": retrieverID,
		"content":     content,
		"key":         key,
		"createdAt":   createdAt,
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
	err = UpdateNotifications(retrieverID, senderID)
	if err != nil {
		return fmt.Errorf("failed to update notifications: %w", err)
	}
	BroadcastNotification(retrieverID, senderID)
	return nil
}
func UpdateNotifications(userID, senderID string) error {
	query := `
		mutation UpdateNotifications($userID: String!, $senderID: String!) {
			insert_notifications(
				objects: {user: $userID, from_users: [$senderID]},
				on_conflict: {constraint: notifications_pkey, update_columns: [from_users]}
			) {
				affected_rows
			}
		}
	`
	variables := map[string]interface{}{
		"userID":   userID,
		"senderID": senderID,
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
	log.Printf("Updated notifications in Hasura")
	return nil
}

func BroadcastMessage(message MessagePusher) {
	pusherID := os.Getenv("PUSHER_APP_ID")
	pusherKey := os.Getenv("PUSHER_APP_KEY")
	pusherSecret := os.Getenv("PUSHER_APP_SECRET")

	pusherClient := pusher.Client{
		AppID:   pusherID,
		Key:     pusherKey,
		Secret:  pusherSecret,
		Cluster: "us2",
		Secure:  true,
	}

	firstID, secondID := message.SenderID, message.RecipientID
	if message.SenderID > message.RecipientID {
		firstID, secondID = message.RecipientID, message.SenderID
	}
	channelName := fmt.Sprintf("private-chat-%s-%s", firstID, secondID)

	data := map[string]interface{}{
		"sender_id":         message.SenderID,
		"recipient_id":      message.RecipientID,
		"encrypted_content": message.EncryptedContent,
		"created_at":        message.CreatedAt,
	}

	err := pusherClient.Trigger(channelName, "new-message", data)
	if err != nil {
		log.Println("Error sending message to Pusher:", err)
	}
}

func BroadcastNotification(userID, senderID string) {
	pusherID := os.Getenv("PUSHER_APP_ID")
	pusherKey := os.Getenv("PUSHER_APP_KEY")
	pusherSecret := os.Getenv("PUSHER_APP_SECRET")

	pusherClient := pusher.Client{
		AppID:   pusherID,
		Key:     pusherKey,
		Secret:  pusherSecret,
		Cluster: "us2",
		Secure:  true,
	}

	channelName := fmt.Sprintf("notifications-%s", userID)

	data := map[string]interface{}{
		"sender_id": senderID,
	}

	err := pusherClient.Trigger(channelName, "new-notification", data)
	if err != nil {
		log.Println("Error sending notification to Pusher:", err)
	}
}
func BroadcastVoiceCall(voicecall VoiceCall) {
	pusherID := os.Getenv("PUSHER_APP_ID")
	pusherKey := os.Getenv("PUSHER_APP_KEY")
	pusherSecret := os.Getenv("PUSHER_APP_SECRET")

	pusherClient := pusher.Client{
		AppID:   pusherID,
		Key:     pusherKey,
		Secret:  pusherSecret,
		Cluster: "us2",
		Secure:  true,
	}
	err := pusherClient.Trigger(
		fmt.Sprintf("private-call-%s", voicecall.CalleeID),
		"incoming-call",
		map[string]interface{}{
			"caller_id": voicecall.CallerID,
			"type":      voicecall.Type,
		},
	)
	if err != nil {
		log.Printf("Error broadcasting call request: %s", err)
	} else {
		log.Printf("Voice call request sent from %s to %s", voicecall.CallerID, voicecall.CalleeID)
	}
}
