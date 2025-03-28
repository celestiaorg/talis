package repos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/celestiaorg/talis/internal/db/models"
)

type JobRepositoryTestSuite struct {
	DBRepositoryTestSuite
}

func TestJobRepository(t *testing.T) {
	suite.Run(t, new(JobRepositoryTestSuite))
}

func (s *JobRepositoryTestSuite) TestCreate() {
	job := s.createTestJob()
	s.NotZero(job.ID)
}

func (s *JobRepositoryTestSuite) TestGetByID() {
	original := s.createTestJob()

	// Test getting with OwnerID
	found, err := s.jobRepo.GetByID(s.ctx, original.OwnerID, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Name, found.Name)

	// Test getting without OwnerID (admin mode)
	found, err = s.jobRepo.GetByID(s.ctx, 0, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)

	// Test with wrong OwnerID
	_, err = s.jobRepo.GetByID(s.ctx, 999, original.ID)
	s.Error(err)

	// Test with non-existent ID
	_, err = s.jobRepo.GetByID(s.ctx, original.OwnerID, 999)
	s.Error(err)
}

func (s *JobRepositoryTestSuite) TestGetByName() {
	job := s.createTestJob()

	// Test getting by name
	found, err := s.jobRepo.GetByName(s.ctx, job.OwnerID, job.Name)
	s.NoError(err)
	s.Equal(job.ID, found.ID)
	s.Equal(job.Name, found.Name)

	// Test getting non-existent job
	_, err = s.jobRepo.GetByName(s.ctx, job.OwnerID, "non-existent")
	s.Error(err)
}

func (s *JobRepositoryTestSuite) TestUpdate() {
	job := s.createTestJob()

	// Update job fields
	job.Status = models.JobStatusCompleted
	err := s.jobRepo.Update(s.ctx, job)
	s.NoError(err)

	// Verify update
	updated, err := s.jobRepo.GetByID(s.ctx, job.OwnerID, job.ID)
	s.NoError(err)
	s.Equal(models.JobStatusCompleted, updated.Status)
}

func (s *JobRepositoryTestSuite) TestUpdateStatus() {
	job := s.createTestJob()

	err := s.jobRepo.UpdateStatus(s.ctx, job.ID, models.JobStatusCompleted, nil, "")
	s.NoError(err)

	updated, err := s.jobRepo.GetByID(s.ctx, job.OwnerID, job.ID)
	s.NoError(err)
	s.Equal(models.JobStatusCompleted, updated.Status)
}

func (s *JobRepositoryTestSuite) TestList() {
	// Create multiple jobs
	s.createTestJob()
	job2 := &models.Job{
		Name:         "test-job-2",
		InstanceName: "test-instance-2",
		ProjectName:  "test-project",
		OwnerID:      2,
		Status:       models.JobStatusPending,
		SSHKeys:      models.SSHKeys{"key1"},
		CreatedAt:    time.Now(),
	}
	s.Require().NoError(s.jobRepo.Create(s.ctx, job2))

	opts := &models.ListOptions{
		Limit:          100,
		Offset:         0,
		IncludeDeleted: false,
	}

	// Test basic listing
	jobs, err := s.jobRepo.List(s.ctx, models.JobStatusUnknown, 0, opts)
	s.NoError(err)
	s.Len(jobs, 2)

	// Test listing with specific status
	jobs, err = s.jobRepo.List(s.ctx, models.JobStatusPending, 0, opts)
	s.NoError(err)
	s.Len(jobs, 2)

	// Test listing with owner ID
	jobs, err = s.jobRepo.List(s.ctx, models.JobStatusUnknown, 1, opts)
	s.NoError(err)
	s.Len(jobs, 1)
}

func (s *JobRepositoryTestSuite) TestCount() {
	// Create multiple jobs
	s.createTestJob()
	s.createTestJob()

	count, err := s.jobRepo.Count(s.ctx, models.JobStatusUnknown, 0)
	s.NoError(err)
	s.Equal(int64(2), count)

	// Test count with specific status
	count, err = s.jobRepo.Count(s.ctx, models.JobStatusPending, 0)
	s.NoError(err)
	s.Equal(int64(2), count)

	// Test count with owner ID
	count, err = s.jobRepo.Count(s.ctx, models.JobStatusUnknown, 1)
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *JobRepositoryTestSuite) TestGetByProjectName() {
	// Create jobs with different project names
	job1 := s.createTestJob()
	job2 := &models.Job{
		Name:         "test-job-2",
		InstanceName: "test-instance-2",
		ProjectName:  "different-project",
		OwnerID:      1,
		Status:       models.JobStatusPending,
		SSHKeys:      models.SSHKeys{"key1"},
		CreatedAt:    time.Now(),
	}
	s.Require().NoError(s.jobRepo.Create(s.ctx, job2))

	// Test getting jobs by project name
	found, err := s.jobRepo.GetByProjectName(s.ctx, job1.ProjectName)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(job1.ID, found.ID)

	// Test getting jobs for non-existent project
	found, err = s.jobRepo.GetByProjectName(s.ctx, "non-existent")
	s.NoError(err)
	s.Nil(found)
}
