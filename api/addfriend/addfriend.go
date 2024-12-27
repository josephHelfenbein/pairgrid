package addfriend

import (
	"api/updateseen"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to add friend to Hasura")

	query := r.URL.Query()
	userID := query.Get("user_id")
	friendEmail := query.Get("friend_email")

	if userID == "" || friendEmail == "" {
		http.Error(w, "Missing user_id or friend_email query parameter", http.StatusBadRequest)
		return
	}
	updateseen.UpdateUserInHasura(userID)
	friendID, err := getFriendIDByEmail(friendEmail)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find friend by email: %s", err), http.StatusInternalServerError)
		log.Printf("Error finding friend by email: %s", err)
		return
	}
	err = insertFriend(userID, friendID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add friend: %s", err), http.StatusInternalServerError)
		log.Printf("Error adding friend: %s", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Friend added successfully"}`))
	log.Printf("Friend added successfully")
}
func getFriendIDByEmail(email string) (string, error) {
	query := `
		query GetUserByEmail($email: String!) {
			users(where: {email: {_eq: $email}}) {
				id
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"email": email,
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request body: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var responseBody struct {
		Data struct {
			Users []struct {
				ID string `json:"id"`
			} `json:"users"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if len(responseBody.Data.Users) == 0 {
		return "", fmt.Errorf("user with email %s not found", email)
	}

	return responseBody.Data.Users[0].ID, nil
}

func insertFriend(userID, friendID string) error {
	firstID, secondID := userID, friendID
	if userID > friendID {
		firstID, secondID = friendID, userID
	}
	if userID == friendID {
		return fmt.Errorf("cannot add self as friend")
	}
	query := `
		query CheckFriendship($first_id: String!, $second_id: String!){
			friends(where: {
				user_id: {_eq: $first_id},
				friend_id: {_eq: $second_id}
			}){
				id
				to_accept
				status	
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"first_id":  firstID,
			"second_id": secondID,
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request body: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var responseBody struct {
		Data struct {
			Friends []struct {
				ID       interface{} `json:"id"`
				ToAccept string      `json:"to_accept"`
				Status   string      `json:"status"`
			} `json:"friends"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if len(responseBody.Data.Friends) > 0 {
		existingFriend := responseBody.Data.Friends[0]
		if existingFriend.Status == "accepted" {
			return fmt.Errorf("friendship already exists")
		}
		if existingFriend.ToAccept == friendID && existingFriend.Status == "pending" {
			return fmt.Errorf("friend request already sent")
		}
		if existingFriend.ToAccept == userID && existingFriend.Status == "pending" {
			mutation := `
				mutation UpdateFriendStatus($id: bigint!, $status: String!) {
					update_friends_by_pk(pk_columns: {id: $id}, _set: {status: $status}) {
						id
					}
				}
			`
			updateRequestBody := map[string]interface{}{
				"query": mutation,
				"variables": map[string]interface{}{
					"id":     existingFriend.ID,
					"status": "accepted",
				},
			}
			jsonBody, err = json.Marshal(updateRequestBody)
			if err != nil {
				return fmt.Errorf("failed to create update request body: %w", err)
			}
			req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				return fmt.Errorf("failed to create update request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-hasura-admin-secret", hasuraSecret)

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to send update request to Hasura: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code for update: %d", resp.StatusCode)
			}
			return nil
		}
	}

	mutation := `
		mutation AddFriend($user_id: String!, $friend_id: String!, $status: String!, $to_accept: String!) {
			insert_friends_one(object: {user_id: $user_id, friend_id: $friend_id, status: $status, to_accept: $to_accept}) {
				id
			}
		}
	`
	insertRequestBody := map[string]interface{}{
		"query": mutation,
		"variables": map[string]interface{}{
			"user_id":   firstID,
			"friend_id": secondID,
			"status":    "pending",
			"to_accept": friendID,
		},
	}
	jsonBody, err = json.Marshal(insertRequestBody)
	if err != nil {
		return fmt.Errorf("failed to create request body: %w", err)
	}

	req, err = http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
