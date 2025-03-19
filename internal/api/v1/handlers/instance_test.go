package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/suite"
)

type InstanceTestSuite struct {
	suite.Suite
	DB           *gorm.DB
	InstanceRepo *repos.InstanceRepository
	JobRepo      *repos.JobRepository
}

func (s *InstanceTestSuite) SetupSuite() {
	var err error
	s.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		s.T().Fatal("failed to connect database")
	}

	// Migrate the schema to create tables
	err = s.DB.AutoMigrate(&models.Instance{}, &models.Job{})
	if err != nil {
		s.T().Fatal("failed to migrate database schema")
	}

	s.InstanceRepo = repos.NewInstanceRepository(s.DB)
	s.JobRepo = repos.NewJobRepository(s.DB)

	s.populateMockDB()
}

func (s *InstanceTestSuite) TearDownSuite() {
	sqlDB, err := s.DB.DB()
	if err == nil {
		err = sqlDB.Close()
		s.NoError(err, "failed to close database connection")
	}
}

func TestInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(InstanceTestSuite))
}

// Helper function to verify the expected output
func (s *InstanceTestSuite) verifyPublicIPsResponse(resp *http.Response, expectedInstances []models.Instance) {
	s.Equal(200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	s.NoError(err)

	expectedMap := make(map[string]models.Instance)
	for _, instance := range expectedInstances {
		expectedMap[instance.PublicIP] = instance
	}

	instances, ok := result["instances"].([]interface{})
	s.True(ok)
	// use greater or equal because terminate test can run concurrently to this test and we can have 3 instances at that point
	s.GreaterOrEqual(len(instances), len(expectedInstances))

	for _, instance := range instances {
		instanceMap := instance.(map[string]interface{})
		publicIP := instanceMap["public_ip"].(string)

		if publicIP == "" {
			continue
		}

		// Access the expected instance using the public IP
		expectedInstance, exists := expectedMap[publicIP]
		s.True(exists, "Expected instance with public IP %s not found", publicIP)

		s.Equal(float64(expectedInstance.JobID), instanceMap["job_id"])
		s.Equal(expectedInstance.PublicIP, instanceMap["public_ip"])
	}
}

func (s *InstanceTestSuite) TestGetPublicIPs() {
	jobService := services.NewJobService(s.JobRepo, s.InstanceRepo)
	instanceService := services.NewInstanceService(s.InstanceRepo, jobService)

	handler := NewInstanceHandler(instanceService)

	app := fiber.New()
	app.Get("/public-ips", handler.GetPublicIPs)

	req := httptest.NewRequest("GET", "/public-ips", nil)
	resp, err := app.Test(req, -1)
	s.NoError(err)

	s.verifyPublicIPsResponse(resp, []models.Instance{
		{
			JobID:    1,
			PublicIP: "192.168.1.1",
		},
		{
			JobID:    2,
			PublicIP: "192.168.1.2",
		},
	})
}

func (s *InstanceTestSuite) TestListInstances() {
	jobService := services.NewJobService(s.JobRepo, s.InstanceRepo)
	instanceService := services.NewInstanceService(s.InstanceRepo, jobService)

	handler := NewInstanceHandler(instanceService)

	app := fiber.New()
	app.Get("/instances", handler.ListInstances)

	req := httptest.NewRequest("GET", "/instances", nil)
	resp, err := app.Test(req)
	s.NoError(err)
	s.Equal(200, resp.StatusCode)
}

func (s *InstanceTestSuite) TestCreateInstance() {

	jobService := services.NewJobService(s.JobRepo, s.InstanceRepo)
	instanceService := services.NewInstanceService(s.InstanceRepo, jobService)

	handler := NewInstanceHandler(instanceService)

	app := fiber.New()
	app.Post("/instances", handler.CreateInstance)

	requestBody := `{"job_name": "test-job", "instances": [{"name": "test-instance", "number_of_instances": 1, "provider": "mock", "region": "nyc1", "size": "s-1vcpu-1gb", "image": "ubuntu-20-04-x64", "tags": ["test"], "ssh_key_name": "default"}]}`
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	s.NoError(err)
	s.Equal(201, resp.StatusCode)
}

func (s *InstanceTestSuite) TestGetInstance() {

	jobService := services.NewJobService(s.JobRepo, s.InstanceRepo)
	instanceService := services.NewInstanceService(s.InstanceRepo, jobService)

	handler := NewInstanceHandler(instanceService)

	app := fiber.New()
	app.Get("/instances/:id", handler.GetInstance)

	req := httptest.NewRequest("GET", "/instances/2", nil)
	resp, err := app.Test(req)
	s.NoError(err)
	s.Equal(200, resp.StatusCode)
}

func (s *InstanceTestSuite) TestTerminateInstance() {
	jobService := services.NewJobService(s.JobRepo, s.InstanceRepo)
	instanceService := services.NewInstanceService(s.InstanceRepo, jobService)

	handler := NewInstanceHandler(instanceService)

	app := fiber.New()
	app.Delete("/instances", handler.TerminateInstances)
	app.Post("/instances", handler.CreateInstance)

	// Create a test instance
	requestBody := `{"job_name": "test-job", "instances": [{"name": "terminate-test-instance", "number_of_instances": 1, "provider": "mock", "region": "nyc1", "size": "s-1vcpu-1gb", "image": "ubuntu-20-04-x64", "tags": ["test"], "ssh_key_name": "default"}]}`
	req := httptest.NewRequest("POST", "/instances", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	s.NoError(err)
	s.Equal(201, resp.StatusCode)

	// TODO: ideally we need to wait for the instance to be created and then terminate it

	// Terminate the test instance
	requestBody = `{"job_name": "test-job", "instance_ids": ["terminate-test-instance"]}`
	req = httptest.NewRequest("DELETE", "/instances", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	s.NoError(err)
	s.Equal(200, resp.StatusCode)
}

func (s *InstanceTestSuite) populateMockDB() {
	instance1 := models.Instance{
		JobID:    1,
		Name:     "test-instance-1",
		PublicIP: "192.168.1.1",
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		Tags:     []string{"test"},
		Status:   models.InstanceStatusReady,
	}
	instance2 := models.Instance{
		JobID:    2,
		Name:     "test-instance-2",
		PublicIP: "192.168.1.2",
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		Tags:     []string{"test"},
		Status:   models.InstanceStatusReady,
	}
	s.NoError(s.InstanceRepo.Create(context.Background(), &instance1))
	s.NoError(s.InstanceRepo.Create(context.Background(), &instance2))

	job := models.Job{
		Name: "test-job",
	}
	err := s.JobRepo.Create(context.Background(), &job)
	s.NoError(err, "failed to create job in mock database")
}
