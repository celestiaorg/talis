package repos

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

type InstanceRepositoryTestSuite struct {
	DBRepositoryTestSuite
}

func TestInstanceRepository(t *testing.T) {
	suite.Run(t, new(InstanceRepositoryTestSuite))
}

func (s *InstanceRepositoryTestSuite) verifyTermination(ownerID, jobID, instanceID uint) error {
	instance, err := s.instanceRepo.GetByID(s.ctx, ownerID, jobID, instanceID)
	if err != nil {
		return err
	}
	if instance.Status != models.InstanceStatusTerminated {
		return fmt.Errorf("expected instance to be terminated: %v", instance.Status)
	}
	// TODO: currently it doesn't seem that we are soft deleting the instances. If this is expected we should remove the delete call in the Terminate method
	if instance.DeletedAt.Valid {
		fmt.Println("expected instance to be deleted")
	}
	return nil
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
	instance2 := s.createTestInstanceForOwner(instance1.OwnerID)

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

	// Update instance
	updateInstance := &models.Instance{
		PublicIP: "192.0.2.100",
		Status:   models.InstanceStatusReady,
	}
	err := s.instanceRepo.UpdateByID(s.ctx, instance.OwnerID, instance.ID, updateInstance)
	s.NoError(err)

	// Verify update
	updated, err := s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.100", updated.PublicIP)
	s.Equal(models.InstanceStatusReady, updated.Status)

	// Test updating IP
	updateIP := &models.Instance{
		PublicIP: "192.0.2.200",
	}
	err = s.instanceRepo.UpdateByName(s.ctx, instance.OwnerID, instance.Name, updateIP)
	s.NoError(err)

	// Verify IP update
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.200", updated.PublicIP)

	// Test updating status
	updateStatus := &models.Instance{
		Status: models.InstanceStatusReady,
	}
	err = s.instanceRepo.UpdateByName(s.ctx, instance.OwnerID, instance.Name, updateStatus)
	s.NoError(err)

	// Verify status update
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal(models.InstanceStatusReady, updated.Status)

	// Test updating multiple fields at once
	updateMultiple := &models.Instance{
		PublicIP: "192.0.2.300",
		Status:   models.InstanceStatusProvisioning,
		Region:   "sfo3",
	}
	err = s.instanceRepo.UpdateByName(s.ctx, instance.OwnerID, instance.Name, updateMultiple)
	s.NoError(err)

	// Verify multiple updates
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.JobID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.300", updated.PublicIP)
	s.Equal(models.InstanceStatusProvisioning, updated.Status)
	s.Equal("sfo3", updated.Region)
}

func (s *InstanceRepositoryTestSuite) TestList() {
	// Create multiple instances with different owners
	instance1 := s.createTestInstance()
	s.createTestInstanceForOwner(instance1.OwnerID)
	s.createTestInstanceForOwner(instance1.OwnerID)

	// Test listing with owner ID
	instancesOwner, err := s.instanceRepo.List(s.ctx, instance1.OwnerID, nil)
	s.NoError(err)
	s.Len(instancesOwner, 3)

	// Test listing with admin ID
	instancesAdmin, err := s.instanceRepo.List(s.ctx, models.AdminID, nil)
	s.NoError(err)
	s.Len(instancesAdmin, 3)

	// Test listing with wrong owner ID
	instancesWrongOwner, err := s.instanceRepo.List(s.ctx, 999, nil)
	s.NoError(err)
	s.Empty(instancesWrongOwner)

	// Test listing with zero owner ID
	_, err = s.instanceRepo.List(s.ctx, 0, nil)
	s.Error(err)
	s.Contains(err.Error(), "invalid owner_id")

	// Test listing with include deleted
	// Terminate an instance and check that it is deleted;
	s.Require().NoError(s.instanceRepo.Terminate(s.ctx, instance1.OwnerID, instance1.ID))
	s.Require().NoError(s.Retry(func() error {
		return s.verifyTermination(instance1.OwnerID, instance1.JobID, instance1.ID)
	}, 50, 100*time.Millisecond))
	instancesDeleted, err := s.instanceRepo.List(s.ctx, instance1.OwnerID, &models.ListOptions{IncludeDeleted: true})
	s.NoError(err)
	s.Len(instancesDeleted, 3)

	instancesNotDeleted, err := s.instanceRepo.List(s.ctx, instance1.OwnerID, &models.ListOptions{IncludeDeleted: false})
	s.NoError(err)
	s.Len(instancesNotDeleted, 2)
}

func (s *InstanceRepositoryTestSuite) TestCount() {
	// Create multiple instances with different owners
	s.createTestInstanceForOwner(1)
	s.createTestInstanceForOwner(1)
	s.createTestInstanceForOwner(1)

	// Test count with owner ID
	count, err := s.instanceRepo.Count(s.ctx, 1)
	s.NoError(err)
	s.Equal(int64(3), count)

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
		OwnerID:    instance1.OwnerID,
		JobID:      instance1.JobID,
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

	// Verify the termination
	s.Require().NoError(s.Retry(func() error {
		return s.verifyTermination(instance.OwnerID, instance.JobID, instance.ID)
	}, 50, 100*time.Millisecond))

	// Test terminate with admin ID
	instance = s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, models.AdminID, instance.ID)
	s.NoError(err)
	s.Require().NoError(s.Retry(func() error {
		return s.verifyTermination(models.AdminID, instance.JobID, instance.ID)
	}, 50, 100*time.Millisecond))

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

func (s *InstanceRepositoryTestSuite) TestApplyListOptions() {
	tests := []struct {
		name     string
		opts     *models.ListOptions
		validate func(query *gorm.DB)
	}{
		{
			name: "nil options",
			opts: nil,
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				s.Contains(sql, "status != 4")
			},
		},
		{
			name: "with status equal filter",
			opts: &models.ListOptions{
				InstanceStatus: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
				StatusFilter: models.StatusFilterEqual,
			},
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				s.Contains(sql, "status = 3")
			},
		},
		{
			name: "with status not equal filter",
			opts: &models.ListOptions{
				InstanceStatus: func() *models.InstanceStatus {
					s := models.InstanceStatusTerminated
					return &s
				}(),
				StatusFilter: models.StatusFilterNotEqual,
			},
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				s.Contains(sql, "status != 4")
			},
		},
		{
			name: "with include deleted",
			opts: &models.ListOptions{
				IncludeDeleted: true,
			},
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				err := query.Find(&instances).Error
				s.NoError(err)
				s.True(query.Statement.Unscoped)
			},
		},
		{
			name: "with pagination",
			opts: &models.ListOptions{
				Limit:  10,
				Offset: 20,
			},
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				s.Contains(sql, "LIMIT 10")
				s.Contains(sql, "OFFSET 20")
			},
		},
		{
			name: "with all options",
			opts: &models.ListOptions{
				Limit:          10,
				Offset:         20,
				IncludeDeleted: true,
				InstanceStatus: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
				StatusFilter: models.StatusFilterEqual,
			},
			validate: func(query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				s.True(query.Statement.Unscoped)
				s.Contains(sql, "LIMIT 10")
				s.Contains(sql, "OFFSET 20")
				s.Contains(sql, "status = 3")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			query := s.instanceRepo.applyListOptions(s.db, tt.opts)
			tt.validate(query)
		})
	}
}
