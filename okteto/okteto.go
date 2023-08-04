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

type OktetoResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []OktetoError          `json:"errors"`
}

type OktetoError struct {
	Message   string           `json:"message"`
	Locations []OktetoLocation `json:"locations"`
	Path      []string         `json:"path"`
}

type OktetoLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (c *Client) NewSecret(name string, value string) error {
	// Define the GraphQL mutation
	mutation := `{"query":"mutation addSecret($name: String!, $value: String!) {\n  addSecret(name: $name, value: $value) {\n    name\n    value\n  }\n}","variables":{"name":"%s","value":"%s"},"operationName":"addSecret"}`
	result, err := c.query(fmt.Sprintf(mutation, name, value))
	if err != nil {
		return err
	}
	// Check if the secret was added successfully
	if result.Data["addSecret"] == nil {
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
	if result.Data["deleteSecret"] == nil {
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
	// Check if the pipeline was scheduled successfully
	if result.Data["deployGitRepository"] == nil {
		fmt.Println("Failed to add pipeline.")
		fmt.Println("Response:", result)
		return err
	}
	fmt.Println("Pipline scheduled successfully!")
	return nil
}

func (c *Client) GetPipeline(namespace string, name string) (map[string]interface{}, error) {
	query := `{"query":"query getSpace($spaceId: String!) {\n  space(id: $spaceId) {\n    id\n    status\n    quotas {\n      ...QuotasFields\n    }\n    members {\n      ...MemberFields\n    }\n    apps {\n      ...AppFields\n    }\n    gitDeploys {\n      ...GitDeployFields\n    }\n    devs {\n      ...DevFields\n    }\n    deployments {\n      ...DeploymentFields\n    }\n    pods {\n      ...PodFields\n    }\n    functions {\n      ...FunctionFields\n    }\n    statefulsets {\n      ...StatefulsetFields\n    }\n    jobs {\n      ...JobFields\n    }\n    cronjobs {\n      ...CronjobFields\n    }\n    volumes {\n      ...VolumeFields\n    }\n    externals {\n      ...ExternalResourceFields\n    }\n    scope\n    persistent\n  }\n}\n\nfragment QuotasFields on Quotas {\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n  pods {\n    ...QuotaFields\n  }\n  storage {\n    ...QuotaFields\n  }\n}\n\nfragment QuotaFields on Resource {\n  limits\n  limitsTotal\n  requests\n  requestsTotal\n  total\n  used\n}\n\nfragment MemberFields on Member {\n  id\n  avatar\n  email\n  externalID\n  name\n  owner\n}\n\nfragment AppFields on App {\n  id\n  name\n  version\n  chart\n  icon\n  description\n  repo\n  config\n  status\n  actionName\n  createdAt\n  updatedAt\n}\n\nfragment GitDeployFields on GitDeploy {\n  id\n  name\n  icon\n  yaml\n  repository\n  repoFullName\n  filename\n  branch\n  status\n  actionName\n  variables {\n    name\n    value\n  }\n  github {\n    installationId\n  }\n  gitCatalogItem {\n    id\n    name\n  }\n  createdAt\n  updatedAt\n}\n\nfragment DevFields on Dev {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  replicas\n  numPods\n  createdAt\n  updatedAt\n  divert\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n  endpoints {\n    ...EndpointFields\n  }\n}\n\nfragment EndpointFields on Endpoint {\n  url\n  private\n  divert\n}\n\nfragment DeploymentFields on Deployment {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  devmode\n  repository\n  path\n  replicas\n  numPods\n  createdAt\n  updatedAt\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n  endpoints {\n    ...EndpointFields\n  }\n}\n\nfragment PodFields on Pod {\n  id\n  name\n  yaml\n  createdAt\n  updatedAt\n  error\n  status\n  deployedBy\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n}\n\nfragment FunctionFields on Function {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  devmode\n  replicas\n  numPods\n  createdAt\n  updatedAt\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n  endpoints {\n    ...EndpointFields\n  }\n}\n\nfragment StatefulsetFields on StatefulSet {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  replicas\n  numPods\n  createdAt\n  updatedAt\n  devmode\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n  endpoints {\n    ...EndpointFields\n  }\n}\n\nfragment JobFields on Job {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  replicas\n  numPods\n  createdAt\n  updatedAt\n  cpu {\n    ...QuotaFields\n  }\n  memory {\n    ...QuotaFields\n  }\n}\n\nfragment CronjobFields on CronJob {\n  id\n  name\n  deployedBy\n  yaml\n  error\n  status\n  createdAt\n  updatedAt\n}\n\nfragment VolumeFields on Volume {\n  id\n  name\n  createdByDevmode\n  deployedBy\n  yaml\n  status\n  createdAt\n  updatedAt\n  storage {\n    ...QuotaFields\n  }\n}\n\nfragment ExternalResourceFields on ExternalResource {\n  id\n  name\n  icon\n  createdAt\n  updatedAt\n  deployedBy\n  endpoints {\n    url\n  }\n  notes {\n    path\n    markdown\n  }\n}","variables":{"spaceId":"%s"},"operationName":"getSpace"}`
	result, err := c.query(fmt.Sprintf(query, namespace))
	if err != nil {
		return nil, err
	}

	space, _ := result.Data["space"].(map[string]interface{})
	gitDeploys, _ := space["gitDeploys"].([]interface{})

	for _, pipeline := range gitDeploys {
		pipelineName, ok := pipeline.(map[string]interface{})["name"].(string)
		if ok && pipelineName == name {
			fmt.Println("Pipeline exists!")
			pipelineData, ok := pipeline.(map[string]interface{})
			if ok {
				return pipelineData, nil
			}
			return nil, fmt.Errorf("could not get pipeline data: %s", pipeline)
		}
	}
	fmt.Println("Pipeline doesn't exist!")
	return nil, nil
}

func (c *Client) DestroyPipeline(name string, namespace string, force bool) error {
	// Define the GraphQL mutation
	mutation := `	{"query":"mutation destroyGitRepository($name: String!, $spaceId: String!, $destroyVolumes: Boolean, $forceDestroy: Boolean) {\n  destroyGitRepository(\n    name: $name\n    space: $spaceId\n    destroyVolumes: $destroyVolumes\n    forceDestroy: $forceDestroy\n  ) {\n    gitDeploy {\n      name\n    }\n    action {\n      status\n    }\n  }\n}","variables":{"name":"%s","spaceId":"%s","destroyVolumes":true,"forceDestroy":%s},"operationName":"destroyGitRepository"}`
	sforce := "false"
	if force {
		sforce = "true"
	}
	result, err := c.query(fmt.Sprintf(mutation, name, namespace, sforce))
	if err != nil {
		return err
	}
	if result.Data["destroyGitRepository"] == nil && result.Errors[0].Message != "not-found" {
		return fmt.Errorf("failed to destroy pipeline: %s", result.Errors[0].Message)
	}
	fmt.Println("Pipeline destroyed successfully!")
	return nil
}

func (c *Client) query(query string) (*OktetoResponse, error) {
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

	var result OktetoResponse

	// b, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(string(b))
	// err = json.Unmarshal(b, &result)

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil, err
	}
	return &result, nil
}
