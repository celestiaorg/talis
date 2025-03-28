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
	instance := s.createTestInstance()
	s.NotZero(instance.ID)
}

func (s *InstanceRepositoryTestSuite) TestGetByID() {
	original := s.createTestInstance()

	// Test getting with JobID
	found, err := s.instanceRepo.GetByID(s.ctx, original.JobID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Name, found.Name)

	// Test getting without JobID (admin mode)
	found, err = s.instanceRepo.GetByID(s.ctx, 0, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test with wrong JobID
	_, err = s.instanceRepo.GetByID(s.ctx, 999, original.ID)
	s.Error(err)

	// Test with non-existent ID
	_, err = s.instanceRepo.GetByID(s.ctx, original.JobID, 999)
	s.Error(err)
}

func (s *InstanceRepositoryTestSuite) TestGetByNames() {
	instance1 := s.createTestInstance()
	instance2 := &models.Instance{
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

	// Test getting multiple instances
	instances, err := s.instanceRepo.GetByNames(s.ctx, []string{instance1.Name, instance2.Name})
	s.NoError(err)
	s.Len(instances, 2)

	// Test getting non-existent instance
	instances, err = s.instanceRepo.GetByNames(s.ctx, []string{"non-existent"})
	s.NoError(err)
	s.Empty(instances)
}

func (s *InstanceRepositoryTestSuite) TestUpdate() {
	instance := s.createTestInstance()

	// Update instance
	instance.PublicIP = "192.0.2.100"
	instance.Status = models.InstanceStatusReady
	err := s.instanceRepo.Update(s.ctx, instance.ID, instance)
	s.NoError(err)

	// Verify update
	updated, err := s.instanceRepo.GetByID(s.ctx, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.100", updated.PublicIP)
	s.Equal(models.InstanceStatusReady, updated.Status)
}

func (s *InstanceRepositoryTestSuite) TestUpdateIPByName() {
	instance := s.createTestInstance()
	newIP := "192.0.2.200"

	err := s.instanceRepo.UpdateIPByName(s.ctx, instance.Name, newIP)
	s.NoError(err)

	updated, err := s.instanceRepo.GetByID(s.ctx, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal(newIP, updated.PublicIP)
}

func (s *InstanceRepositoryTestSuite) TestUpdateStatus() {
	instance := s.createTestInstance()

	err := s.instanceRepo.UpdateStatus(s.ctx, instance.ID, models.InstanceStatusReady)
	s.NoError(err)

	updated, err := s.instanceRepo.GetByID(s.ctx, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal(models.InstanceStatusReady, updated.Status)
}

func (s *InstanceRepositoryTestSuite) TestUpdateStatusByName() {
	instance := s.createTestInstance()

	err := s.instanceRepo.UpdateStatusByName(s.ctx, instance.Name, models.InstanceStatusReady)
	s.NoError(err)

	updated, err := s.instanceRepo.GetByID(s.ctx, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal(models.InstanceStatusReady, updated.Status)
}

func (s *InstanceRepositoryTestSuite) TestList() {
	// Create multiple instances
	s.createTestInstance()
	s.createTestInstance()

	// Test basic listing
	instances, err := s.instanceRepo.List(s.ctx, nil)
	s.NoError(err)
	s.Len(instances, 2)

	// Test listing with include deleted
	instance := s.createTestInstance()
	s.Require().NoError(s.instanceRepo.Terminate(s.ctx, instance.ID))

	instances, err = s.instanceRepo.List(s.ctx, &models.ListOptions{IncludeDeleted: true})
	s.NoError(err)
	s.Len(instances, 3)

	instances, err = s.instanceRepo.List(s.ctx, &models.ListOptions{IncludeDeleted: false})
	s.NoError(err)
	s.Len(instances, 2)
}

func (s *InstanceRepositoryTestSuite) TestCount() {
	// Create multiple instances
	s.createTestInstance()
	s.createTestInstance()

	count, err := s.instanceRepo.Count(s.ctx)
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *InstanceRepositoryTestSuite) TestGetByJobID() {
	// Create instances with different job IDs
	instance1 := s.createTestInstance()
	instance2 := &models.Instance{
		JobID:      2,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting instances for job 1
	instances, err := s.instanceRepo.GetByJobID(s.ctx, 1)
	s.NoError(err)
	s.Len(instances, 1)
	s.Equal(instance1.ID, instances[0].ID)

	// Test getting instances for non-existent job
	instances, err = s.instanceRepo.GetByJobID(s.ctx, 999)
	s.NoError(err)
	s.Empty(instances)
}

func (s *InstanceRepositoryTestSuite) TestGetByJobIDOrdered() {
	// Create instances with different creation times
	instance1 := s.createTestInstance()
	time.Sleep(time.Millisecond) // Ensure different creation times
	instance2 := &models.Instance{
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-2",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance2))

	// Test getting ordered instances
	instances, err := s.instanceRepo.GetByJobIDOrdered(s.ctx, 1)
	s.NoError(err)
	s.Len(instances, 2)
	s.Equal(instance1.ID, instances[0].ID)
	s.Equal(instance2.ID, instances[1].ID)
}

func (s *InstanceRepositoryTestSuite) TestTerminate() {
	instance := s.createTestInstance()

	// Test termination
	err := s.instanceRepo.Terminate(s.ctx, instance.ID)
	s.NoError(err)

	// Verify instance is soft deleted and status is updated
	instances, err := s.instanceRepo.List(s.ctx, &models.ListOptions{IncludeDeleted: true})
	s.NoError(err)
	found := false
	for _, i := range instances {
		if i.ID == instance.ID {
			found = true
			s.Equal(models.InstanceStatusTerminated, i.Status)
			s.NotNil(i.DeletedAt)
		}
	}
	s.True(found, "Terminated instance should be found when including deleted")

	// Verify instance is not in regular list
	instances, err = s.instanceRepo.List(s.ctx, nil)
	s.NoError(err)
	for _, i := range instances {
		s.NotEqual(instance.ID, i.ID, "Terminated instance should not be in regular list")
	}
}
