package pusherauth

import (
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
	log.Printf("Parsing channel name: %s", channelName)
	if !strings.HasPrefix(channelName, "private-chat-") {
		if !strings.HasPrefix(channelName, "private-call-") {
			return "", "", fmt.Errorf("invalid channel name prefix")
		}
		trimmedChannelName := channelName[len("private-call-"):]
		parts := strings.SplitN(trimmedChannelName, "-", 1)
		if len(parts) != 1 {
			return "", "", fmt.Errorf("unexpected channel name format, expected one ID")
		}
		return parts[0], "", nil
	}
	trimmedChannelName := channelName[len("private-chat-"):]
	parts := strings.SplitN(trimmedChannelName, "-", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected channel name format, expected two IDs")
	}
	return parts[0], parts[1], nil
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
	params := []byte(r.Form.Encode())
	log.Printf("Query params: %s", r.Form.Encode())

	authResponse, err := pusherClient.AuthorizePrivateChannel(params)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		log.Printf("Pusher private channel authorization failed: %v", err)
		return
	}
	log.Printf("Authorization response: %s", authResponse)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", string(authResponse))
	log.Printf("Users successfully authenticated")
}
