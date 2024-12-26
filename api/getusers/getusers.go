package getusers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type User struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Bio        string   `json:"bio"`
	Language   []string `json:"language"`
	Specialty  string   `json:"specialty"`
	Interests  []string `json:"interests"`
	Occupation string   `json:"occupation"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get user from Hasura")
	offset := 0
	limit := 10
	userID := ""
	query := r.URL.Query()
	if o := query.Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}
	if u := query.Get("user_id"); u != "" {
		userID = u
	}
	users, err := GetUsersFromHasura(offset, limit, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user from Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error getting user from Hasura: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	}
	log.Printf("Users successfully retrieved from Hasura")
}
func GetFriendLists(userID string) ([]string, []string, error) {
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
		return nil, nil, fmt.Errorf("failed to create request body: %w", err)
	}
	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}
	friendIDs := make([]string, len(responseBody.Data.Friends1))
	for i, f := range responseBody.Data.Friends1 {
		friendIDs[i] = f.FriendID
	}
	userIDs := make([]string, len(responseBody.Data.Friends2))
	for i, f := range responseBody.Data.Friends2 {
		userIDs[i] = f.UserID
	}
	return friendIDs, userIDs, nil
}
func GetUsersFromHasura(offset, limit int, userID string) ([]User, error) {
	friendIDs, userIDs, err := GetFriendLists(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend lists: %w", err)
	}
	query := `
		query GetUsers($offset: Int!, $limit: Int!, $userID: String!, $friendIDs: [String!], $userIDs: [String!]) {
			users(where: {
				_and: [
					{id: {_neq: $userID}},
					{id: {_nin: $friendIDs}},
					{id: {_nin: $userIDs}}
				]
			}, offset: $offset, limit: $limit) {
				name
				email
				bio
				language
				specialty
				interests
				occupation
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"offset":    offset,
			"limit":     limit,
			"userID":    userID,
			"friendIDs": friendIDs,
			"userIDs":   userIDs,
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
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	log.Printf("Hasura response: %+v", responseBody.Data.Users)
	return responseBody.Data.Users, nil
}
