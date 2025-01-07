package pusherauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/pusher/pusher-http-go/v5"
)

var pusherClient = pusher.Client{
	AppID:   os.Getenv("PUSHER_APP_ID"),
	Key:     os.Getenv("PUSHER_APP_KEY"),
	Secret:  os.Getenv("PUSHER_APP_SECRET"),
	Cluster: "us2",
	Secure:  true,
}

func parseChannelName(channelName string) (string, string, error) {
	var firstID, secondID string
	_, err := fmt.Sscanf(channelName, "chat-%s-%s", &firstID, &secondID)
	if err != nil {
		return "", "", err
	}
	return firstID, secondID, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
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

	params, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Printf("Error reading request body: %v", err)
		return
	}
	var requestData struct {
		ChannelName string `json:"channel_name"`
		SocketID    string `json:"socket_id"`
	}
	if err := json.Unmarshal(params, &requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error parsing request body: %v", err)
		return
	}

	firstID, secondID, err := parseChannelName(requestData.ChannelName)
	if err != nil {
		http.Error(w, "Invalid channel name", http.StatusBadRequest)
		log.Printf("Channel name parsing failed: %v", err)
		return
	}
	if usr.ID != firstID && usr.ID != secondID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("User %s is not authorized to access channel %s", usr.ID, requestData.ChannelName)
		return
	}
	presenceData := pusher.MemberData{
		UserID: usr.ID,
		UserInfo: map[string]string{
			"id": usr.ID,
		},
	}
	authResponse, err := pusherClient.AuthorizePresenceChannel(params, presenceData)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		log.Printf("Pusher presence channel authorization failed: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse)
}
