package okteto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	apiURL = "https://cloud.okteto.com/graphql"
)

type Client struct {
	Namespace  string
	BaseURL    *url.URL
	HTTPClient *http.Client

	apiToken string
}

// NewClient creates new Okteto client.
func NewClient(apiToken string, namespace string) *Client {
	c := &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		apiToken:   apiToken,
		Namespace:  namespace,
	}
	c.BaseURL, _ = url.Parse(apiURL)
	return c
}

type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SecretResponse struct {
	ID string `json:"id"`
}

func (c *Client) NewSecret(name string, value string) (*string, error) {
	// Define the GraphQL mutation
	mutation := `
	mutation {
		addSecret(input: {
			name: "%s"
			value: "%s"
		}) {
			name
			value
		}
	}`

	// Format the GraphQL mutation with the secret data
	query := fmt.Sprintf(mutation, name, value)

	// Prepare the API request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(query))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the API request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check the API response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to add secret: %s\n", resp.Status)
		return nil, err
	}

	// Parse the API response
	var result map[string]map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil, err
	}

	// Check if the secret was added successfully
	addedSecret := result["data"]["addSecret"]
	if addedSecret != nil {
		fmt.Println("Secret added successfully!")
		fmt.Printf("Name: %s\n", addedSecret["name"])
		fmt.Printf("Value: %s\n", addedSecret["value"])
	} else {
		fmt.Println("Failed to add secret.")
		fmt.Println("Response:", result)
	}

	// secret := Secret{
	// 	Name:  name,
	// 	Value: value,
	// }

	// secretJSON, err := json.Marshal(secret)
	// if err != nil {
	// 	fmt.Println("Error marshaling secret:", err)
	// 	return nil, err
	// }

	// url := fmt.Sprintf("%s/user/secrets", c.BaseURL.String())
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(secretJSON))
	// if err != nil {
	// 	fmt.Println("Error creating request:", err)
	// 	return nil, err
	// }

	// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	// req.Header.Set("Content-Type", "application/json")

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error sending request:", err)
	// 	return nil, err
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusCreated {
	// 	fmt.Printf("Failed to add secret. Status code: %d\n", resp.StatusCode)
	// 	return nil, err
	// }

	// var secretResp SecretResponse
	// err = json.NewDecoder(resp.Body).Decode(&secretResp)
	// if err != nil {
	// 	fmt.Println("Error decoding response", err)
	// 	return nil, err
	// }

	// fmt.Println("Secret added successfully!")
	// return &secretResp.ID, nil
	return nil, nil
}
