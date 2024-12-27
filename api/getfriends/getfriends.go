package getfriends

import (
	"api/updateseen"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type User struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profile_picture"`
	LastSeen       string `json:"last_seen"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get friends from Hasura")

	query := r.URL.Query()
	userID := query.Get("user_id")

	if userID == "" {
		http.Error(w, "Missing user_id query parameter", http.StatusBadRequest)
		return
	}
	updateseen.UpdateUserInHasura(userID)
	friendLists, err := GetFriendLists(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get friends from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting friends from Hasura: %s", err)
		return
	}
	users, err := GetUsersInfo(friendLists)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get users info from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting users info from Hasura: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	log.Printf("Friends successfully retrieved from Hasura")
}
func GetFriendLists(userID string) ([]string, error) {
	query := `
		query GetFriends($userID: String!) {
			friends1: friends(where: {user_id: {_eq: $userID}, status: {_eq: "accepted"}}) {
				friend_id
			}
			friends2: friends(where: {friend_id: {_eq: $userID}, status: {_eq: "accepted"}}) {
				user_id
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"userID": userID,
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

	client := &http.Client{}
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
			Friends1 []struct {
				FriendID string `json:"friend_id"`
			} `json:"friends1"`
			Friends2 []struct {
				UserID string `json:"user_id"`
			} `json:"friends2"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	friendList := []string{}
	for _, friend := range responseBody.Data.Friends1 {
		if friend.FriendID != userID {
			friendList = append(friendList, friend.FriendID)
		}
	}
	for _, friend := range responseBody.Data.Friends2 {
		if friend.UserID != userID {
			friendList = append(friendList, friend.UserID)
		}
	}
	return friendList, nil
}
func GetUsersInfo(userIDs []string) ([]User, error) {
	if len(userIDs) == 0 {
		return []User{}, nil
	}
	query := `
		query GetUsersInfo($userIDs: [String!]!) {
			users(where: {id: {_in: $userIDs}}) {
				name
				email
				profile_picture
				last_seen
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"userIDs": userIDs,
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

	client := &http.Client{}
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
			Users []User `json:"users"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return responseBody.Data.Users, nil
}
