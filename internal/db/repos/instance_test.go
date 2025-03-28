package repos

import (
	"testing"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the test database
	err = db.AutoMigrate(&models.Instance{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestInstanceRepository_applyListOptions(t *testing.T) {
	db, err := setupTestDB()
	require.NoError(t, err)

	repo := NewInstanceRepository(db)

	tests := []struct {
		name     string
		opts     *models.ListOptions
		validate func(t *testing.T, query *gorm.DB)
	}{
		{
			name: "nil options",
			opts: nil,
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				assert.Contains(t, sql, "status != 4")
			},
		},
		{
			name: "with status equal filter",
			opts: &models.ListOptions{
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
				StatusFilter: models.StatusFilterEqual,
			},
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				assert.Contains(t, sql, "status = 3")
			},
		},
		{
			name: "with status not equal filter",
			opts: &models.ListOptions{
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusTerminated
					return &s
				}(),
				StatusFilter: models.StatusFilterNotEqual,
			},
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				assert.Contains(t, sql, "status != 4")
			},
		},
		{
			name: "with include deleted",
			opts: &models.ListOptions{
				IncludeDeleted: true,
			},
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				err := query.Find(&instances).Error
				require.NoError(t, err)
				assert.True(t, query.Statement.Unscoped)
			},
		},
		{
			name: "with pagination",
			opts: &models.ListOptions{
				Limit:  10,
				Offset: 20,
			},
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				assert.Contains(t, sql, "LIMIT 10")
				assert.Contains(t, sql, "OFFSET 20")
			},
		},
		{
			name: "with all options",
			opts: &models.ListOptions{
				Limit:          10,
				Offset:         20,
				IncludeDeleted: true,
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
				StatusFilter: models.StatusFilterEqual,
			},
			validate: func(t *testing.T, query *gorm.DB) {
				var instances []models.Instance
				sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
					return tx.Find(&instances)
				})
				assert.True(t, query.Statement.Unscoped)
				assert.Contains(t, sql, "LIMIT 10")
				assert.Contains(t, sql, "OFFSET 20")
				assert.Contains(t, sql, "status = 3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := repo.applyListOptions(db, tt.opts)
			tt.validate(t, query)
		})
	}
}
