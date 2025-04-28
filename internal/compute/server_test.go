package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestServerService_Create(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var serverResp Server
		serverResp.ID = 1
		serverResp.Name = "test-server"
		serverResp.Memory = 2048
		serverResp.Disk = 50
		serverResp.CPU = 2
		serverResp.Status = "pending"

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(serverResp)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	createReq := &ServerCreateRequest{
		Name:             "test-server",
		Memory:           2048,
		Disk:             50,
		CPU:              2,
		PackageID:        1,
		FirewallRulesets: []int{1},
	}

	createdServer, apiResp, err := serverService.Create(context.Background(), createReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, apiResp.StatusCode)
	assert.NotNil(t, createdServer)
	assert.Equal(t, createReq.Name, createdServer.Name)
}

func TestServerService_Get(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers/1", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		var serverResp Server
		serverResp.ID = 1
		serverResp.Name = "test-server"
		serverResp.Status = "running"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(serverResp)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	retrievedServer, apiResp, err := serverService.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)
	assert.NotNil(t, retrievedServer)
	assert.Equal(t, int(1), retrievedServer.ID)
}

func TestServerService_Delete(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers/1", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	apiResp, err := serverService.Delete(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, apiResp.StatusCode)
}

func TestServerService_List(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		var result struct {
			Servers []Server `json:"servers"`
		}
		result.Servers = []Server{
			{
				ID:     1,
				Name:   "test-server-1",
				Status: "running",
			},
			{
				ID:     2,
				Name:   "test-server-2",
				Status: "stopped",
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	servers, apiResp, err := serverService.List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)
	assert.NotNil(t, servers)
	assert.Len(t, servers, 2)
}

func TestServerService_Build(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers/1/build", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	apiResp, err := serverService.Build(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, apiResp.StatusCode)
}

func TestServerService_Suspend(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers/1/suspend", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	apiResp, err := serverService.Suspend(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, apiResp.StatusCode)
}

func TestServerService_Unsuspend(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/servers/1/unsuspend", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer testServer.Close()

	cfg := &config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     testServer.URL,
	}
	client, err := NewClient(cfg)
	assert.NoError(t, err)

	serverService := newServerService(client)

	apiResp, err := serverService.Unsuspend(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, apiResp.StatusCode)
}
