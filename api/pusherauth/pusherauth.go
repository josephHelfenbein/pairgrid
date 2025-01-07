package pusherauth

import (
	"encoding/json"
	"fmt"
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
	_, err := fmt.Sscanf(channelName, "private-chat-%s-%s", &firstID, &secondID)
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

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		log.Printf("Error parsing form data: %v", err)
		return
	}
	channelName := r.FormValue("channel_name")
	socketID := r.FormValue("socket_id")

	if channelName == "" || socketID == "" {
		http.Error(w, "Missing required form fields", http.StatusBadRequest)
		log.Printf("Missing channel_name or socket_id")
		return
	}

	firstID, secondID, err := parseChannelName(channelName)
	if err != nil {
		http.Error(w, "Invalid channel name", http.StatusBadRequest)
		log.Printf("Channel name parsing failed: %v", err)
		return
	}
	if usr.ID != firstID && usr.ID != secondID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("User %s is not authorized to access channel %s", usr.ID, channelName)
		return
	}
	presenceData := pusher.MemberData{
		UserID: usr.ID,
		UserInfo: map[string]string{
			"id": usr.ID,
		},
	}
	requestData := struct {
		ChannelName string `json:"channel_name"`
		SocketID    string `json:"socket_id"`
	}{
		ChannelName: channelName,
		SocketID:    socketID,
	}
	paramsJSON, err := json.Marshal(requestData)
	if err != nil {
		http.Error(w, "Error marshalling request data", http.StatusInternalServerError)
		log.Printf("Error marshalling request data: %v", err)
		return
	}
	authResponse, err := pusherClient.AuthorizePresenceChannel(paramsJSON, presenceData)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		log.Printf("Pusher presence channel authorization failed: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse)
}
