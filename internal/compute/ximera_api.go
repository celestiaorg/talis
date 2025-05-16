package compute

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	computeTypes "github.com/celestiaorg/talis/internal/compute/types"
)

// XimeraAPIClient is a client for interacting with the Ximera API
// Uses local ximera models
type XimeraAPIClient struct {
	config     *computeTypes.Configuration
	httpClient *http.Client
}

// NewXimeraAPIClient creates a new API client
func NewXimeraAPIClient(config *computeTypes.Configuration) *XimeraAPIClient {
	return &XimeraAPIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MakeRequest makes a request to the API
func (c *XimeraAPIClient) MakeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", c.config.APIURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		cerr := resp.Body.Close()
		if cerr != nil {
			fmt.Printf("warning: error closing response body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

// ListServers lists all servers
func (c *XimeraAPIClient) ListServers() (*computeTypes.ServersListResponse, error) {
	respBody, err := c.MakeRequest("GET", "/servers", nil)
	if err != nil {
		return nil, err
	}

	var response computeTypes.ServersListResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// ServerExists checks if a server with the given name exists
func (c *XimeraAPIClient) ServerExists(name string) (bool, int, error) {
	servers, err := c.ListServers()
	if err != nil {
		return false, 0, err
	}

	for _, server := range servers.Data {
		if server.Name == name {
			return true, server.ID, nil
		}
	}

	return false, 0, nil
}

// CreateServer creates a new server
func (c *XimeraAPIClient) CreateServer(name string, packageID int, storage, traffic, memory, cpuCores int) (*computeTypes.ServerResponse, error) {
	request := computeTypes.ServerCreateRequest{
		PackageID:    packageID,
		UserID:       c.config.UserID,
		HypervisorID: c.config.HypervisorID,
		IPv4:         1,
		Name:         name,
	}

	// Add optional parameters if provided
	if storage > 0 {
		request.Storage = storage
	}
	if traffic > 0 {
		request.Traffic = traffic
	}
	if memory > 0 {
		request.Memory = memory
	}
	if cpuCores > 0 {
		request.CPUCores = cpuCores
	}

	respBody, err := c.MakeRequest("POST", "/servers", request)
	if err != nil {
		return nil, err
	}

	var response computeTypes.ServerResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// GetServer gets a server by ID
func (c *XimeraAPIClient) GetServer(id int) (*computeTypes.ServerResponse, error) {
	endpoint := fmt.Sprintf("/servers/%d", id)
	respBody, err := c.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response computeTypes.ServerResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Extract public IP address from the full API response
	var fullResponse map[string]interface{}
	err = json.Unmarshal(respBody, &fullResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling full response: %w", err)
	}

	// Navigate through the JSON structure to find the public IP
	if data, ok := fullResponse["data"].(map[string]interface{}); ok {
		if network, ok := data["network"].(map[string]interface{}); ok {
			if interfaces, ok := network["interfaces"].([]interface{}); ok && len(interfaces) > 0 {
				if iface, ok := interfaces[0].(map[string]interface{}); ok {
					if ipv4, ok := iface["ipv4"].([]interface{}); ok && len(ipv4) > 0 {
						if ip, ok := ipv4[0].(map[string]interface{}); ok {
							if address, ok := ip["address"].(string); ok {
								response.Data.PublicIP = address
							}
						}
					}
				}
			}
		}
	}

	return &response, nil
}

// BuildServer builds a server with the given ID
func (c *XimeraAPIClient) BuildServer(id int, osID, name, sshKey string) (*computeTypes.ServerResponse, error) {
	osIDInt, err := strconv.Atoi(osID)
	if err != nil {
		return nil, fmt.Errorf("error converting osID to int: %w", err)
	}
	sshKeyInt, err := strconv.Atoi(sshKey)
	if err != nil {
		return nil, fmt.Errorf("error converting sshKey to int: %w", err)
	}

	request := computeTypes.ServerBuildRequest{
		OperatingSystemID: osIDInt,
		Name:              name,
		Hostname:          "",
		SSHKeys:           []int{sshKeyInt},
	}

	endpoint := fmt.Sprintf("/servers/%d/build", id)
	respBody, err := c.MakeRequest("POST", endpoint, request)
	if err != nil {
		return nil, err
	}

	var response computeTypes.ServerResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// ListTemplates lists available OS templates for a package
func (c *XimeraAPIClient) ListTemplates(packageID int) (*computeTypes.TemplatesResponse, error) {
	endpoint := fmt.Sprintf("/media/templates/fromServerPackageSpec/%d", packageID)
	respBody, err := c.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response computeTypes.TemplatesResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &response, nil
}

// DeleteServer deletes a server with the given ID
func (c *XimeraAPIClient) DeleteServer(id int) error {
	endpoint := fmt.Sprintf("/servers/%d", id)
	_, err := c.MakeRequest("DELETE", endpoint, nil)
	return err
}

// WaitForServerCreation waits for a server to be fully created
func (c *XimeraAPIClient) WaitForServerCreation(serverID int, timeoutSeconds int) error {
	fmt.Printf("Waiting for server creation to complete...")

	interval := 5 * time.Second
	maxWait := time.Duration(timeoutSeconds) * time.Second
	start := time.Now()

	for {
		server, err := c.GetServer(serverID)
		if err != nil {
			return fmt.Errorf("error getting server details: %w", err)
		}
		if server != nil && (server.Data.State == "complete") {
			fmt.Printf(" done (state: %s)\n", server.Data.State)
			return nil
		}
		if time.Since(start) > maxWait {
			return fmt.Errorf("timeout waiting for server to be running (last state: %s)", server.Data.State)
		}
		fmt.Printf(". ")
		time.Sleep(interval)
	}
}
