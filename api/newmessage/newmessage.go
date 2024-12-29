package newmessage

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pusher/pusher-http-go/v5"
)

type HasuraEvent struct {
	Payload struct {
		New Message `json:"new"`
	} `json:"payload"`
	Version string `json:"version"`
}

type Message struct {
	ID               string `json:"id"`
	SenderID         string `json:"sender_id"`
	RecipientID      string `json:"recipient_id"`
	EncryptedContent string `json:"encrypted_content"`
	Key              string `json:"key"`
	CreatedAt        string `json:"created_at"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var event HasuraEvent
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode request body: %s", err), http.StatusBadRequest)
		return
	}
	log.Printf("Decoded HasuraEvent: %+v", event)

	message := event.Payload.New
	log.Printf("Extracted Message: %+v", message)

	BroadcastMessage(message)
}

func BroadcastMessage(message Message) {
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
	channelName := fmt.Sprintf("chat-%s-%s", firstID, secondID)
	log.Printf("Message ID: %s", message.ID)
	log.Printf("Sender ID: %s", message.SenderID)
	log.Printf("Recipient ID: %s", message.RecipientID)
	log.Printf("Encrypted Content: %s", message.EncryptedContent)
	log.Printf("Key: %s", message.Key)
	log.Printf("Created At: %s", message.CreatedAt)
	data := map[string]interface{}{
		"id":                message.ID,
		"sender_id":         message.SenderID,
		"recipient_id":      message.RecipientID,
		"encrypted_content": message.EncryptedContent,
		"key":               message.Key,
		"created_at":        message.CreatedAt,
	}

	err := pusherClient.Trigger(channelName, "new-message", data)
	if err != nil {
		log.Println("Error sending message to Pusher:", err)
	}
}
