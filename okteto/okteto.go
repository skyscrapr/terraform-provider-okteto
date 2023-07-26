package okteto

import (
	"bytes"
	"encoding/json"
	// "io"
	// "log"
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

func (c *Client) NewSecret(name string, value string) error {
	// Define the GraphQL mutation
	mutation := `{"query":"mutation addSecret($name: String!, $value: String!) {\n  addSecret(name: $name, value: $value) {\n    name\n    value\n  }\n}","variables":{"name":"%s","value":"%s"},"operationName":"addSecret"}`
	result, err := c.query(fmt.Sprintf(mutation, name, value))
	if err != nil {
		return err
	}
	// Check if the secret was added successfully
	if result["data"]["addSecret"] == nil {
		fmt.Println("Failed to add secret.")
		fmt.Println("Response:", result)
		return err
	}
	fmt.Println("Secret added successfully!")
	return nil
}

// func (c *Client) GetSecret(name string) error {
// 	// Define the GraphQL mutation
// 	query := `{"query":"query fetchUserAndTeam {\n  user {\n    ...UserFields\n  }\n  team {\n    id\n    name\n    avatar\n    members {\n      id\n      avatar\n      email\n      externalID\n      name\n      owner\n    }\n  }\n}\n\nfragment UserFields on me {\n  id\n  externalID\n  githubID\n  namespace\n  avatar\n  name\n  email\n  token\n  new\n  super\n  config\n  userDefinedNamespaces\n  plan\n  planSubscribed\n  trialExpiration\n  quotaPlan {\n    maxNamespaces\n    maxPods\n    scaleToZeroPeriod\n    enableSharing\n    enableCustomCatalog\n    enablePlans\n    limits {\n      cpu\n      memory\n      storage\n    }\n    limitRanges {\n      max {\n        cpu\n        memory\n      }\n    }\n  }\n  secrets {\n    name\n    value\n  }\n  personalAccessTokens {\n    id\n    name\n    expirationDate\n    status\n  }\n  secondaryEmails\n  quickstarts {\n    name\n    url\n    branch\n    default\n    variables {\n      name\n      value\n      options\n    }\n  }\n  integrations {\n    github {\n      enabled\n      url\n      connected\n      appInstallationUrl\n      authUrl\n    }\n  }\n  capabilities {\n    maxPersonalAccessTokens\n    teamsEnabled\n    automaticPreviewsEnabled\n    newOnboardingEnabled\n    helmCatalogEnabled\n    allowMembersShareNamespace\n    shareNamespaceOnlyWithUsersEnabled\n    namespacesPrefix\n    userNamespacesSuffix\n    installationBoardEnabled\n  }\n  team\n}","operationName":"fetchUserAndTeam"}`
// 	result, err := c.query(query)
// 	if err != nil {
// 		return err
// 	}
// 	// Check if the secret exists
// 	secrets := result["data"]["user"]["data"]["secrets"]["data"]
// 	for secret, i := range(secrets) {
// 		if secret["data"]["name"] == name {
// 			fmt.Println("Secret exists!")
// 			return  nil
// 		}
// 	}
// 	fmt.Println("Secret doesn't exist!")
// 	return  err
// }

func (c *Client) DeleteSecret(name string) error {
	// Define the GraphQL mutation
	mutation := `{"query":"mutation deleteSecret($name: String!) {\n  deleteSecret(name: $name) {\n    name\n    value\n  }\n}","variables":{"name":"%s"},"operationName":"deleteSecret"}`
	result, err := c.query(fmt.Sprintf(mutation, name))
	if err != nil {
		return err
	}
	// Check if the secret was added successfully
	if result["data"]["deleteSecret"] == nil {
		fmt.Println("Failed to delete secret.")
		fmt.Println("Response:", result)
		return err
	}
	fmt.Println("Secret deleted successfully!")
	return nil
}

func (c *Client) NewPipeline(namespace string, name string, repo string, branch string) error {
	// Define the GraphQL mutation
	mutation := `{"query":"mutation deployGitRepository($name: String!, $space: String!, $source: String!, $branch: String, $repository: String!, $installationId: String, $variables: [InputVariable], $filename: String, $catalogItemId: String) {\n  deployGitRepository(\n    name: $name\n    space: $space\n    source: $source\n    branch: $branch\n    repository: $repository\n    installationId: $installationId\n    variables: $variables\n    filename: $filename\n    catalogItemId: $catalogItemId\n  ) {\n    gitDeploy {\n      id\n      status\n    }\n    action {\n      status\n    }\n  }\n}","variables":{"space":"%s","name":"%s","repository":"%s","branch":"%s","variables":[],"filename":"","source":"ui","catalogItemId":null},"operationName":"deployGitRepository"}`
	// skyscrapr
	// 39784259
	// okteto-aws-lambda
	// https://github.com/skyscrapr/okteto-aws-lambda.git
	// main

	result, err := c.query(fmt.Sprintf(mutation, namespace, name, repo, branch))
	if err != nil {
		return err
	}
	// Check if the secret was added successfully
	if result["data"]["deployGitRepository"] == nil {
		fmt.Println("Failed to add pipeline.")
		fmt.Println("Response:", result)
		return err
	}
	fmt.Println("Pipline scheduled successfully!")
	return nil
}

func (c *Client) DestroyPipeline(name string, namespace string) error {
	// Define the GraphQL mutation
	mutation := `	{"query":"mutation destroyGitRepository($name: String!, $spaceId: String!, $destroyVolumes: Boolean, $forceDestroy: Boolean) {\n  destroyGitRepository(\n    name: $name\n    space: $spaceId\n    destroyVolumes: $destroyVolumes\n    forceDestroy: $forceDestroy\n  ) {\n    gitDeploy {\n      name\n    }\n    action {\n      status\n    }\n  }\n}","variables":{"name":"%s","spaceId":"%s","destroyVolumes":true,"forceDestroy":true},"operationName":"destroyGitRepository"}`
	result, err := c.query(fmt.Sprintf(mutation, name, namespace))
	if err != nil {
		return err
	}
	// Check if the secret was added successfully
	if result["data"]["destroyGitRepository"] == nil {
		fmt.Println("Failed to destroy pipeline.")
		fmt.Println("Response:", result)
		return err
	}
	fmt.Println("Pipeline destroyed successfully!")
	return nil
}

func (c *Client) query(query string) (map[string]map[string]interface{}, error) {
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
		fmt.Printf("Failed to execute query: %s\n", resp.Status)
		return nil, err
	}

	// // Parse the API response
	// b, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(string(b))
	// return nil, nil

	var result map[string]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil, err
	}
	return result, nil
}
