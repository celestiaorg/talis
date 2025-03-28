package repos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/celestiaorg/talis/internal/db/models"
)

type InstanceRepositoryTestSuite struct {
	DBRepositoryTestSuite
}

func TestInstanceRepository(t *testing.T) {
	suite.Run(t, new(InstanceRepositoryTestSuite))
}

func (s *InstanceRepositoryTestSuite) TestCreate() {
	// Test with valid owner ID
	instance := s.createTestInstance()
	s.NotZero(instance.ID)

	// Test with zero owner ID should fail
	invalidInstance := &models.Instance{
		OwnerID:    0,
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-invalid",
	}
	err := s.instanceRepo.Create(s.ctx, invalidInstance)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestGetByID() {
	original := s.createTestInstance()

	// Test getting with correct owner ID
	found, err := s.instanceRepo.GetByID(s.ctx, original.OwnerID, original.JobID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Name, found.Name)

	// Test getting with admin ID
	found, err = s.instanceRepo.GetByID(s.ctx, models.AdminID, original.JobID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test getting with wrong owner ID
	_, err = s.instanceRepo.GetByID(s.ctx, 999, original.JobID, original.ID)
	s.Error(err)

	// Test getting with zero owner ID
	_, err = s.instanceRepo.GetByID(s.ctx, 0, original.JobID, original.ID)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")

	// Test with non-existent ID
	_, err = s.instanceRepo.GetByID(s.ctx, original.OwnerID, original.JobID, 999)
	s.Error(err)
}

func (s *InstanceRepositoryTestSuite) TestGetByNames() {
	instance1 := s.createTestInstance()
	instance2 := &models.Instance{
		OwnerID:    1,
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		PublicIP:   "192.0.2.2",
		Region:     "nyc1",
		Size:       "s-1vcpu-1gb",
		Image:      "ubuntu-20-04-x64",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting with correct owner ID
	instances, err := s.instanceRepo.GetByNames(s.ctx, instance1.OwnerID, []string{instance1.Name, instance2.Name})
	s.NoError(err)
	s.Len(instances, 2)

	// Test getting with admin ID
	instances, err = s.instanceRepo.GetByNames(s.ctx, models.AdminID, []string{instance1.Name, instance2.Name})
	s.NoError(err)
	s.Len(instances, 2)

	// Test getting with wrong owner ID
	instances, err = s.instanceRepo.GetByNames(s.ctx, 999, []string{instance1.Name, instance2.Name})
	s.NoError(err)
	s.Empty(instances)

	// Test getting with zero owner ID
	_, err = s.instanceRepo.GetByNames(s.ctx, 0, []string{instance1.Name})
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestUpdate() {
	instance := s.createTestInstance()

	// Test update with correct owner ID
	instance.PublicIP = "192.0.2.100"
	instance.Status = models.InstanceStatusReady
	err := s.instanceRepo.Update(s.ctx, instance.OwnerID, instance.ID, instance)
	s.NoError(err)

	// Verify update
	updated, err := s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.100", updated.PublicIP)
	s.Equal(models.InstanceStatusReady, updated.Status)

	// Test update with admin ID
	instance.PublicIP = "192.0.2.101"
	err = s.instanceRepo.Update(s.ctx, models.AdminID, instance.ID, instance)
	s.NoError(err)

	// Test update with wrong owner ID
	err = s.instanceRepo.Update(s.ctx, 999, instance.ID, instance)
	s.Error(err)
	s.Contains(err.Error(), "instance not found or not owned by user")

	// Test update with zero owner ID
	err = s.instanceRepo.Update(s.ctx, 0, instance.ID, instance)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestUpdateIPByName() {
	instance := s.createTestInstance()
	newIP := "192.0.2.200"

	// Test update with correct owner ID
	err := s.instanceRepo.UpdateIPByName(s.ctx, instance.OwnerID, instance.Name, newIP)
	s.NoError(err)

	// Test update with admin ID
	err = s.instanceRepo.UpdateIPByName(s.ctx, models.AdminID, instance.Name, newIP)
	s.NoError(err)

	// Test update with wrong owner ID
	err = s.instanceRepo.UpdateIPByName(s.ctx, 999, instance.Name, newIP)
	s.Error(err)
	s.Contains(err.Error(), "instance not found or not owned by user")

	// Test update with zero owner ID
	err = s.instanceRepo.UpdateIPByName(s.ctx, 0, instance.Name, newIP)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestUpdateStatus() {
	instance := s.createTestInstance()

	// Test update with correct owner ID
	err := s.instanceRepo.UpdateStatus(s.ctx, instance.OwnerID, instance.ID, models.InstanceStatusReady)
	s.NoError(err)

	// Test update with admin ID
	err = s.instanceRepo.UpdateStatus(s.ctx, models.AdminID, instance.ID, models.InstanceStatusReady)
	s.NoError(err)

	// Test update with wrong owner ID
	err = s.instanceRepo.UpdateStatus(s.ctx, 999, instance.ID, models.InstanceStatusReady)
	s.Error(err)
	s.Contains(err.Error(), "instance not found or not owned by user")

	// Test update with zero owner ID
	err = s.instanceRepo.UpdateStatus(s.ctx, 0, instance.ID, models.InstanceStatusReady)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestUpdateStatusByName() {
	instance := s.createTestInstance()

	// Test update with correct owner ID
	err := s.instanceRepo.UpdateStatusByName(s.ctx, instance.OwnerID, instance.Name, models.InstanceStatusReady)
	s.NoError(err)

	// Test update with admin ID
	err = s.instanceRepo.UpdateStatusByName(s.ctx, models.AdminID, instance.Name, models.InstanceStatusReady)
	s.NoError(err)

	// Test update with wrong owner ID
	err = s.instanceRepo.UpdateStatusByName(s.ctx, 999, instance.Name, models.InstanceStatusReady)
	s.Error(err)
	s.Contains(err.Error(), "instance not found or not owned by user")

	// Test update with zero owner ID
	err = s.instanceRepo.UpdateStatusByName(s.ctx, 0, instance.Name, models.InstanceStatusReady)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestList() {
	// Create multiple instances with different owners
	s.createTestInstance()
	s.createTestInstance()
	s.createTestInstance()

	// Test listing with owner ID
	instances, err := s.instanceRepo.List(s.ctx, 1, nil)
	s.NoError(err)
	s.Len(instances, 2)

	// Test listing with admin ID
	instances, err = s.instanceRepo.List(s.ctx, models.AdminID, nil)
	s.NoError(err)
	s.Len(instances, 3)

	// Test listing with wrong owner ID
	instances, err = s.instanceRepo.List(s.ctx, 999, nil)
	s.NoError(err)
	s.Empty(instances)

	// Test listing with zero owner ID
	_, err = s.instanceRepo.List(s.ctx, 0, nil)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")

	// Test listing with include deleted
	instance := s.createTestInstance()
	s.Require().NoError(s.instanceRepo.Terminate(s.ctx, instance.OwnerID, instance.ID))

	instances, err = s.instanceRepo.List(s.ctx, 1, &models.ListOptions{IncludeDeleted: true})
	s.NoError(err)
	s.Len(instances, 3)

	instances, err = s.instanceRepo.List(s.ctx, 1, &models.ListOptions{IncludeDeleted: false})
	s.NoError(err)
	s.Len(instances, 2)
}

func (s *InstanceRepositoryTestSuite) TestCount() {
	// Create multiple instances with different owners
	s.createTestInstance()
	s.createTestInstance()
	s.createTestInstance()

	// Test count with owner ID
	count, err := s.instanceRepo.Count(s.ctx, 1)
	s.NoError(err)
	s.Equal(int64(2), count)

	// Test count with admin ID
	count, err = s.instanceRepo.Count(s.ctx, models.AdminID)
	s.NoError(err)
	s.Equal(int64(3), count)

	// Test count with wrong owner ID
	count, err = s.instanceRepo.Count(s.ctx, 999)
	s.NoError(err)
	s.Equal(int64(0), count)

	// Test count with zero owner ID
	_, err = s.instanceRepo.Count(s.ctx, 0)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestGetByJobID() {
	// Create instances with different job IDs and owners
	instance1 := s.createTestInstance()
	instance2 := &models.Instance{
		OwnerID:    2,
		JobID:      2,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting with correct owner ID
	instances, err := s.instanceRepo.GetByJobID(s.ctx, instance1.OwnerID, instance1.JobID)
	s.NoError(err)
	s.Len(instances, 1)
	s.Equal(instance1.ID, instances[0].ID)

	// Test getting with admin ID
	instances, err = s.instanceRepo.GetByJobID(s.ctx, models.AdminID, instance1.JobID)
	s.NoError(err)
	s.Len(instances, 1)

	// Test getting with wrong owner ID
	instances, err = s.instanceRepo.GetByJobID(s.ctx, 999, instance1.JobID)
	s.NoError(err)
	s.Empty(instances)

	// Test getting with zero owner ID
	_, err = s.instanceRepo.GetByJobID(s.ctx, 0, instance1.JobID)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestGetByJobIDOrdered() {
	// Create instances with different creation times and owners
	instance1 := s.createTestInstance()
	time.Sleep(time.Millisecond) // Ensure different creation times
	instance2 := &models.Instance{
		OwnerID:    1,
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting ordered instances with correct owner ID
	instances, err := s.instanceRepo.GetByJobIDOrdered(s.ctx, instance1.OwnerID, instance1.JobID)
	s.NoError(err)
	s.Len(instances, 2)
	s.Equal(instance1.ID, instances[0].ID)
	s.Equal(instance2.ID, instances[1].ID)

	// Test getting ordered instances with admin ID
	instances, err = s.instanceRepo.GetByJobIDOrdered(s.ctx, models.AdminID, instance1.JobID)
	s.NoError(err)
	s.Len(instances, 2)

	// Test getting ordered instances with wrong owner ID
	instances, err = s.instanceRepo.GetByJobIDOrdered(s.ctx, 999, instance1.JobID)
	s.NoError(err)
	s.Empty(instances)

	// Test getting ordered instances with zero owner ID
	_, err = s.instanceRepo.GetByJobIDOrdered(s.ctx, 0, instance1.JobID)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestTerminate() {
	instance := s.createTestInstance()

	// Test terminate with correct owner ID
	err := s.instanceRepo.Terminate(s.ctx, instance.OwnerID, instance.ID)
	s.NoError(err)

	// Test terminate with admin ID
	instance = s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, models.AdminID, instance.ID)
	s.NoError(err)

	// Test terminate with wrong owner ID
	instance = s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, 999, instance.ID)
	s.Error(err)
	s.Contains(err.Error(), "instance not found or not owned by user")

	// Test terminate with zero owner ID
	instance = s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, 0, instance.ID)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestQuery() {
	// Create test instances
	s.createTestInstance()
	s.createTestInstance()

	// Test query with admin ID
	instances, err := s.instanceRepo.Query(s.ctx, models.AdminID, "SELECT * FROM instances")
	s.NoError(err)
	s.Len(instances, 2)

	// Test query with non-admin ID should fail
	_, err = s.instanceRepo.Query(s.ctx, 1, "SELECT * FROM instances")
	s.Error(err)
	s.Contains(err.Error(), "restricted to admin users only")

	// Test query with zero owner ID
	_, err = s.instanceRepo.Query(s.ctx, 0, "SELECT * FROM instances")
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}

func (s *InstanceRepositoryTestSuite) TestGetByJobIDAndNames() {
	// Create test instances with different owners
	instance1 := s.createTestInstance()
	instance2 := &models.Instance{
		OwnerID:    2,
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting with correct owner ID
	instances, err := s.instanceRepo.GetByJobIDAndNames(s.ctx, instance1.OwnerID, instance1.JobID, []string{instance1.Name})
	s.NoError(err)
	s.Len(instances, 1)
	s.Equal(instance1.ID, instances[0].ID)

	// Test getting with admin ID
	instances, err = s.instanceRepo.GetByJobIDAndNames(s.ctx, models.AdminID, instance1.JobID, []string{instance1.Name, instance2.Name})
	s.NoError(err)
	s.Len(instances, 2)

	// Test getting with wrong owner ID
	instances, err = s.instanceRepo.GetByJobIDAndNames(s.ctx, 999, instance1.JobID, []string{instance1.Name})
	s.NoError(err)
	s.Empty(instances)

	// Test getting with zero owner ID
	_, err = s.instanceRepo.GetByJobIDAndNames(s.ctx, 0, instance1.JobID, []string{instance1.Name})
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")
}
