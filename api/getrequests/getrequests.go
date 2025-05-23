package getrequests

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
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	ProfilePicture string   `json:"profile_picture"`
	Bio            string   `json:"bio"`
	Language       []string `json:"language"`
	Specialty      string   `json:"specialty"`
	Interests      []string `json:"interests"`
	Occupation     string   `json:"occupation"`
	LastSeen       string   `json:"last_seen"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get friends from Hasura")

	query := r.URL.Query()
	userID := query.Get("user_id")
	kind := query.Get("kind")

	if userID == "" || kind == "" {
		http.Error(w, "Missing one or more query parameters", http.StatusBadRequest)
		return
	}
	friendLists := []string(nil)
	err := error(nil)
	if kind == "friend" {
		friendLists, err = GetFriendLists(userID)
	} else if kind == "request" {
		friendLists, err = GetRequestLists(userID)
	} else if kind == "notifications" {
		friendLists, err = GetNotifications(userID)
	} else {
		http.Error(w, "Invalid kind query parameter", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get friends from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting friends from Hasura: %s", err)
		return
	}
	if kind == "notifications" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(friendLists); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
			log.Printf("Error creating response JSON: %s", err)
			return
		}
		log.Printf("Notifications successfully retrieved from Hasura")
		return
	}
	updateseen.UpdateUserInHasura(userID)
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
func GetRequestLists(userID string) ([]string, error) {
	query := `
		query GetFriends($userID: String!) {
			friends1: friends(where: {user_id: {_eq: $userID}, status: {_eq: "pending"}, to_accept: {_eq: $userID}}) {
				friend_id
			}
			friends2: friends(where: {friend_id: {_eq: $userID}, status: {_eq: "pending"}, to_accept: {_eq: $userID}}) {
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
func GetNotifications(userID string) ([]string, error) {
	query := `
		query GetNotifications($userID: String!) {
			notifications(where: {user: {_eq: $userID}}) {
				from_users
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
			Notifications []struct {
				FromUsers []string `json:"from_users"`
			} `json:"notifications"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	var fromUsers []string
	for _, notification := range responseBody.Data.Notifications {
		fromUsers = append(fromUsers, notification.FromUsers...)
	}

	return fromUsers, nil
}
func GetUsersInfo(userIDs []string) ([]User, error) {
	if len(userIDs) == 0 {
		return []User{}, nil
	}
	query := `
		query GetUsersInfo($userIDs: [String!]!) {
			users(where: {id: {_in: $userIDs}}) {
				id
				name
				email
				bio
				language
				specialty
				interests
				occupation
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
