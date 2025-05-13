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
	instance, err := s.instanceRepo.Get(s.ctx, ownerID, instanceID)
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
	_, err := s.instanceRepo.Create(s.ctx, invalidInstance)
	s.NoError(err) // Temporarily allowing zero owner_id

	// Test CreateBatch
	instances := []*models.Instance{
		s.randomInstance(),
		s.randomInstance(),
	}
	_, err = s.instanceRepo.CreateBatch(s.ctx, instances)
	s.NoError(err)

	// Confirm all instances were created
	dbInstances, err := s.instanceRepo.List(s.ctx, models.AdminID, nil)
	s.NoError(err)
	s.Len(dbInstances, 3)
}

func (s *InstanceRepositoryTestSuite) TestGet() {
	original := s.createTestInstance()

	// Test getting with correct owner ID
	found, err := s.instanceRepo.Get(s.ctx, original.OwnerID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test getting with admin ID
	found, err = s.instanceRepo.Get(s.ctx, models.AdminID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test getting with wrong owner ID
	wrongOwnerID := original.OwnerID + 1
	if wrongOwnerID == models.AdminID { // Edge case: if original.OwnerID + 1 is AdminID
		wrongOwnerID++
	}
	if wrongOwnerID == 0 { // Edge case: if original.OwnerID was MaxUint - 1, then +1 could wrap to 0 (if uint)
		wrongOwnerID = original.OwnerID - 1 // or some other distinct non-zero, non-admin value
	}
	_, err = s.instanceRepo.Get(s.ctx, wrongOwnerID, original.ID)
	s.Error(err)

	// Test getting with zero owner ID should work but log a warning
	// TODO: Once ValidateOwnerID returns an error, update this test to expect an error
	found, err = s.instanceRepo.Get(s.ctx, 0, original.ID)
	s.NoError(err)
	s.NotNil(found)

	// Test with non-existent ID
	nonExistentID := original.ID + 1
	// Ensure nonExistentID is truly non-existent, could also use a very large number
	// For simplicity, original.ID + 1 is usually fine if DB is clean for the test.
	_, err = s.instanceRepo.Get(s.ctx, original.OwnerID, nonExistentID)
	s.Error(err)
}

func (s *InstanceRepositoryTestSuite) TestUpdate() {
	instance := s.createTestInstance()

	// Update instance
	updateInstance := &models.Instance{
		PublicIP: "192.0.2.100",
		Status:   models.InstanceStatusReady,
		Region:   "sfo2",
	}
	err := s.instanceRepo.Update(s.ctx, instance.OwnerID, instance.ID, updateInstance)
	s.NoError(err)

	// Verify update
	updated, err := s.instanceRepo.Get(s.ctx, instance.OwnerID, instance.ID)
	s.NoError(err)
	s.Equal("192.0.2.100", updated.PublicIP)
	s.Equal(models.InstanceStatusReady, updated.Status)
	s.Equal("sfo2", updated.Region)

	// Check that other fields were not modified
	s.Equal(instance.ID, updated.ID)
	s.Equal(instance.OwnerID, updated.OwnerID)
	s.Equal(instance.ProjectID, updated.ProjectID)
	s.Equal(instance.ProviderID, updated.ProviderID)
	s.Equal(instance.Size, updated.Size)
	s.Equal(instance.Image, updated.Image)
	s.Equal(instance.Tags, updated.Tags)
	s.Equal(instance.CreatedAt.Unix(), updated.CreatedAt.Unix())
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

	// Test query with valid query by ID
	instances, err := s.instanceRepo.Query(s.ctx, instance1.OwnerID, models.InstanceIDField+" = ?", instance1.ID)
	s.NoError(err)
	s.Len(instances, 1)
	s.Equal(instance1.ID, instances[0].ID)

	// Test query with admin ID for multiple IDs
	instances, err = s.instanceRepo.Query(s.ctx, models.AdminID, models.InstanceIDField+" = ? OR "+models.InstanceIDField+" = ?", instance1.ID, instance2.ID)
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

func (s *InstanceRepositoryTestSuite) TestGetByProjectIDAndInstanceIDs() {
	ownerID := s.randomOwnerID()
	project1 := s.createTestProjectForOwner(ownerID)
	project2 := s.createTestProjectForOwner(ownerID) // Another project for the same owner

	instance1P1 := s.createTestInstanceForOwnerAndProject(ownerID, project1.ID)
	instance2P1 := s.createTestInstanceForOwnerAndProject(ownerID, project1.ID)
	instance1P2 := s.createTestInstanceForOwnerAndProject(ownerID, project2.ID)

	// Test case 1: Get instances from project1 by their IDs
	foundP1, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID, project1.ID, []uint{instance1P1.ID, instance2P1.ID})
	s.NoError(err)
	s.Len(foundP1, 2)
	idsP1 := []uint{foundP1[0].ID, foundP1[1].ID}
	s.Contains(idsP1, instance1P1.ID)
	s.Contains(idsP1, instance2P1.ID)

	// Test case 2: Try to get an instance from project1 using an ID from project2 (should not be found for project1)
	partialFoundP1, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID, project1.ID, []uint{instance1P1.ID, instance1P2.ID})
	s.NoError(err)
	s.Len(partialFoundP1, 1)
	s.Equal(instance1P1.ID, partialFoundP1[0].ID)

	// Test case 3: Get an instance from project2
	foundP2, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID, project2.ID, []uint{instance1P2.ID})
	s.NoError(err)
	s.Len(foundP2, 1)
	s.Equal(instance1P2.ID, foundP2[0].ID)

	// Test case 4: Admin can get instances regardless of direct ownership if project matches
	adminFoundP1, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, models.AdminID, project1.ID, []uint{instance1P1.ID, instance2P1.ID})
	s.NoError(err)
	s.Len(adminFoundP1, 2)

	// Test case 5: No IDs provided
	noIDs, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID, project1.ID, []uint{})
	s.NoError(err) // As per current implementation, returns empty slice and no error
	s.Empty(noIDs)

	// Test case 6: Non-existent IDs for a project
	nonExistentIDs, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID, project1.ID, []uint{9998, 9999})
	s.NoError(err)
	s.Empty(nonExistentIDs)

	// Test case 7: Wrong owner ID (but correct project ID which implies ownership by someone else)
	// The method filters by ownerID first if not admin, then by projectID and then by instanceIDs.
	// So, if ownerID doesn't match, it should return empty even if projectID and instanceIDs are valid for another owner.
	wrongOwner, err := s.instanceRepo.GetByProjectIDAndInstanceIDs(s.ctx, ownerID+100, project1.ID, []uint{instance1P1.ID})
	s.NoError(err)
	s.Empty(wrongOwner)
}

// Helper needed for TestGetByProjectIDAndInstanceIDs
func (s *DBRepositoryTestSuite) createTestInstanceForOwnerAndProject(ownerID, projectID uint) *models.Instance {
	instance := &models.Instance{
		OwnerID:    ownerID,
		ProjectID:  projectID,
		ProviderID: models.ProviderDO,
		PublicIP:   fmt.Sprintf("192.0.2.%d", projectID), // Make it somewhat unique for test
		Region:     "nyc1",
		Size:       "s-1vcpu-1gb",
		Image:      "ubuntu-20-04-x64",
		Tags:       []string{"test", "dev"},
		Status:     models.InstanceStatusPending,
		CreatedAt:  time.Now(),
	}
	_, err := s.instanceRepo.Create(s.ctx, instance)
	s.Require().NoError(err)
	return instance
}
