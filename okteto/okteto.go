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
	apiURL = "https://cloud.okteto.com/api"
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
	Name  string            `json:"name"`
	Value map[string]string `json:"value"`
}

type SecretResponse struct {
	ID string `json:"id"`
}

func (c *Client) NewSecret(name string, value map[string]string) (*string, error) {
	secret := Secret{
		Name:  name,
		Value: value,
	}

	secretJSON, err := json.Marshal(secret)
	if err != nil {
		fmt.Println("Error marshaling secret:", err)
		return nil, err
	}

	url := fmt.Sprintf("%s/namespaces/%s/secrets", c.BaseURL.String(), c.Namespace)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(secretJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Failed to add secret. Status code: %d\n", resp.StatusCode)
		return nil, err
	}

	var secretResp SecretResponse
	err = json.NewDecoder(resp.Body).Decode(&secretResp)
	if err != nil {
		fmt.Println("Error decoding response", err)
		return nil, err
	}

	fmt.Println("Secret added successfully!")
	return &secretResp.ID, nil
}
