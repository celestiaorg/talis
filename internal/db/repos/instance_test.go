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

func (s *InstanceRepositoryTestSuite) verifyTermination(ownerID, instanceID uint) error {
	instance, err := s.instanceRepo.GetByID(s.ctx, ownerID, instanceID)
	if err != nil {
		return err
	}
	if instance.Status != models.InstanceStatusTerminated {
		return fmt.Errorf("instance is not yet terminated (status=%v)", instance.Status)
	}
	return nil
}

func (s *InstanceRepositoryTestSuite) TestCreate() {
	// Test with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	invalidInstance := s.randomInstanceForOwner(0)
	err := s.instanceRepo.Create(s.ctx, invalidInstance)
	s.NoError(err) // Temporarily allowing zero owner_id

	// Test CreateBatch
	instances := []*models.Instance{
		s.randomInstance(),
		s.randomInstance(),
	}
	err = s.instanceRepo.CreateBatch(s.ctx, instances)
	s.NoError(err)

	// Confirm all instances were created
	dbInstances, err := s.instanceRepo.List(s.ctx, models.AdminID, nil)
	s.NoError(err)
	s.Len(dbInstances, 3)
}

func (s *InstanceRepositoryTestSuite) TestGetByID() {
	original := s.createTestInstance()

	// Test getting with correct owner ID
	found, err := s.instanceRepo.GetByID(s.ctx, original.OwnerID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Name, found.Name)

	// Test getting with admin ID
	found, err = s.instanceRepo.GetByID(s.ctx, models.AdminID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test getting with wrong owner ID
	_, err = s.instanceRepo.GetByID(s.ctx, 999, original.ID)
	s.Error(err)

	// Test getting with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	found, err = s.instanceRepo.GetByID(s.ctx, 0, original.ID)
	s.NoError(err)
	s.NotNil(found)

	// Test with non-existent ID
	_, err = s.instanceRepo.GetByID(s.ctx, original.OwnerID, 999)
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

	// Test getting with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	instances, err = s.instanceRepo.GetByNames(s.ctx, 0, []string{instance1.Name})
	s.NoError(err)
	s.NotNil(instances)
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
	updated, err := s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.ID)
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
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.200", updated.PublicIP)

	// Test updating status
	updateStatus := &models.Instance{
		Status: models.InstanceStatusReady,
	}
	err = s.instanceRepo.UpdateByName(s.ctx, instance.OwnerID, instance.Name, updateStatus)
	s.NoError(err)

	// Verify status update
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.ID)
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
	updated, err = s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.ID)
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

	// Test listing with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	instancesZeroOwner, err := s.instanceRepo.List(s.ctx, 0, nil)
	s.NoError(err)
	s.NotNil(instancesZeroOwner) // Verify we got a response

	// Test listing with include deleted
	// Terminate an instance and check that it is deleted;
	s.Require().NoError(s.instanceRepo.Terminate(s.ctx, instance1.OwnerID, instance1.ID))
	s.Require().NoError(s.Retry(func() error {
		return s.verifyTermination(instance1.OwnerID, instance1.ID)
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

	// Test count with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	count, err = s.instanceRepo.Count(s.ctx, 0)
	s.NoError(err)
	s.GreaterOrEqual(count, int64(0))
}

func (s *InstanceRepositoryTestSuite) TestQuery() {
	// Create test instances
	instance1 := s.createTestInstance()
	instance2 := s.createTestInstance()

	// Test query with valid query
	instances, err := s.instanceRepo.Query(s.ctx, instance1.OwnerID, "name = ?", instance1.Name)
	s.NoError(err)
	s.Len(instances, 1)
	s.Equal(instance1.ID, instances[0].ID)

	// Test query with admin ID
	instances, err = s.instanceRepo.Query(s.ctx, models.AdminID, "name = ? OR name = ?", instance1.Name, instance2.Name)
	s.NoError(err)
	s.Len(instances, 2)

	// Test query with invalid query
	_, err = s.instanceRepo.Query(s.ctx, instance1.OwnerID, "INVALID SQL")
	s.Error(err)
	s.Contains(err.Error(), "failed to query instances")
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
				s.Contains(sql, "status != 5")
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
				s.Contains(sql, "status = 4")
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
				s.Contains(sql, "status != 5")
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
				s.Contains(sql, "status = 4")
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

func (s *InstanceRepositoryTestSuite) TestTerminate() {
	instance := s.createTestInstance()

	// Test Terminate with correct owner ID
	err := s.instanceRepo.Terminate(s.ctx, instance.OwnerID, instance.ID)
	s.Require().NoError(err)

	// Verify terminated
	err = s.verifyTermination(instance.OwnerID, instance.ID)
	s.NoError(err)

	// Test with wrong owner ID
	instance2 := s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, 999, instance2.ID)
	s.NoError(err) // This is a no-op since it won't find an instance

	// Test with admin ID
	instance3 := s.createTestInstance()
	err = s.instanceRepo.Terminate(s.ctx, models.AdminID, instance3.ID)
	s.NoError(err)
	err = s.verifyTermination(models.AdminID, instance3.ID)
	s.NoError(err)
}

func (s *InstanceRepositoryTestSuite) TestUpdateByName() {
	instance := &models.Instance{
		OwnerID:    1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance-update-by-name",
		Status:     models.InstanceStatusPending,
	}
	s.Require().NoError(s.instanceRepo.Create(s.ctx, instance))

	// Update the instance
	err := s.instanceRepo.UpdateByName(s.ctx, instance.OwnerID, instance.Name, &models.Instance{
		PublicIP: "192.0.2.100",
	})
	s.NoError(err)

	// Verify update
	updated, err := s.instanceRepo.GetByID(s.ctx, instance.OwnerID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.100", updated.PublicIP)
}
