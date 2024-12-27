package deletefriend

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
	err = deleteFriend(userID, friendID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove friend: %s", err), http.StatusInternalServerError)
		log.Printf("Error removing friend: %s", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Friend removed successfully"}`))
	log.Printf("Friend removed successfully")
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

func deleteFriend(userID, friendID string) error {
	firstID, secondID := userID, friendID
	if userID > friendID {
		firstID, secondID = friendID, userID
	}
	if userID == friendID {
		return fmt.Errorf("cannot add self as friend")
	}
	mutation := `
		mutation DeleteFriendship($first_id: String!, $second_id: String!){
			delete_friends(where: {
				user_id: {_eq: $first_id},
				friend_id: {_eq: $second_id}
			}){
				affected_rows
			}
		}
	`
	requestBody := map[string]interface{}{
		"query": mutation,
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
			DeleteFriends struct {
				AffectedRows int `json:"affected_rows"`
			} `json:"delete_friends"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if responseBody.Data.DeleteFriends.AffectedRows == 0 {
		return fmt.Errorf("no friendship row found to delete")
	}

	return nil
}
