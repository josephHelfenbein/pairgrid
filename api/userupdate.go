package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type ClerkUser struct {
	ID               string            `json:"id"`
	FirstName        string            `json:"first_name"`
	LastName         string            `json:"last_name"`
	EmailAddresses   []EmailAddress    `json:"email_addresses"`
	ExternalAccounts []ExternalAccount `json:"external_accounts"`
	ImageURL         string            `json:"image_url"`
	LastActiveAt     int64             `json:"last_active_at"`
	LastSignInAt     *int64            `json:"last_sign_in_at,omitempty"`
	Locked           bool              `json:"locked"`
	Username         string            `json:"username"`
}

type EmailAddress struct {
	EmailAddress string `json:"email_address"`
	ID           string `json:"id"`
	Verification struct {
		Status string `json:"status"`
	} `json:"verification"`
}

type ExternalAccount struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received Clerk webhook request")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %s", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Raw webhook payload: %s", string(body))
	var rawPayload struct {
		Data struct {
			ID               string            `json:"id"`
			FirstName        string            `json:"first_name"`
			LastName         string            `json:"last_name"`
			EmailAddresses   []EmailAddress    `json:"email_addresses"`
			ExternalAccounts []ExternalAccount `json:"external_accounts"`
			ImageURL         string            `json:"image_url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &rawPayload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON payload: %s", err), http.StatusBadRequest)
		log.Printf("Error unmarshalling JSON: %s", err)
		return
	}
	user := ClerkUser{
		ID:               rawPayload.Data.ID,
		FirstName:        rawPayload.Data.FirstName,
		LastName:         rawPayload.Data.LastName,
		EmailAddresses:   rawPayload.Data.EmailAddresses,
		ExternalAccounts: rawPayload.Data.ExternalAccounts,
		ImageURL:         rawPayload.Data.ImageURL,
	}
	log.Printf("Received user: %+v", user)

	if err := CreateUserInHasura(user); err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert user into Hasura: %s", err), http.StatusInternalServerError)
		log.Printf("Error inserting user into Hasura: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "success"}
	if jsonResp, err := json.Marshal(resp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create response JSON: %s", err), http.StatusInternalServerError)
		log.Printf("Error creating response JSON: %s", err)
		return
	} else {
		w.Write(jsonResp)
	}

	log.Printf("User with ID %s successfully created in Hasura", user.ID)
}

func CreateUserInHasura(user ClerkUser) error {
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	var email string
	if len(user.EmailAddresses) > 0 {
		email = user.EmailAddresses[0].EmailAddress
	} else {
		return fmt.Errorf("user does not have any email addresses")
	}
	checkUserQuery := `
        query CheckUser($id: String!) {
            users_by_pk(id: $id) {
                id
            }
        }
    `

	checkVariables := map[string]interface{}{
		"id": user.ID,
	}

	checkRequestBody := map[string]interface{}{
		"query":     checkUserQuery,
		"variables": checkVariables,
	}
	if userExists, err := sendHasuraRequest(checkRequestBody); err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	} else if userExists {
		updateUserMutation := `
            mutation UpdateUser($id: String!, $name: String!, $email: String!, $profile_picture: String!) {
                update_users_by_pk(pk_columns: {id: $id}, _set: {name: $name, email: $email, profile_picture: $profile_picture}) {
                    id
                    name
                    email
                    profile_picture
                }
            }
        `
		updateVariables := map[string]interface{}{
			"id":              user.ID,
			"name":            fullName,
			"email":           email,
			"profile_picture": user.ImageURL,
		}

		updateRequestBody := map[string]interface{}{
			"query":     updateUserMutation,
			"variables": updateVariables,
		}

		if _, err := sendHasuraRequest(updateRequestBody); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		log.Printf("User with ID %s successfully updated in Hasura", user.ID)
	} else {
		insertUserMutation := `
			mutation InsertUsers($id: String!, $name: String!, $email: String!, $profile_picture: String!) {
				insert_users(objects: {id: $id, name: $name, email: $email, profile_picture: $profile_picture}) {
					affected_rows
					returning {
						id
						name
						email
						bio
						language
						specialty
						interests
						occupation
						last_seen
						created_at
						last_typed
						profile_picture
					}
				}
			}
		`

		insertVariables := map[string]interface{}{
			"id":              user.ID,
			"name":            fullName,
			"email":           email,
			"profile_picture": user.ImageURL,
		}

		insertRequestBody := map[string]interface{}{
			"query":     insertUserMutation,
			"variables": insertVariables,
		}

		if _, err := sendHasuraRequest(insertRequestBody); err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		log.Printf("User with ID %s successfully inserted into Hasura", user.ID)
	}
	return nil
}

func sendHasuraRequest(requestBody map[string]interface{}) (bool, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal Hasura request body: %w", err)
	}

	hasuraURL := os.Getenv("HASURA_GRAPHQL_URL")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create Hasura request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request to Hasura: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("hasura responded with status: %s, body: %s", resp.Status, string(body))
	}

	var responseBody struct {
		Data   map[string]interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return false, fmt.Errorf("failed to decode Hasura response: %w", err)
	}

	if len(responseBody.Errors) > 0 {
		return false, fmt.Errorf("hasura returned errors: %v", responseBody.Errors)
	}

	if users, ok := responseBody.Data["users_by_pk"].(map[string]interface{}); ok {
		if users["id"] != nil {
			return true, nil
		}
	}

	return false, nil
}
