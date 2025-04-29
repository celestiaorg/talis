package types

import "time"

// UploadDeletionPayload defines the data for a delete_upload task
type UploadDeletionPayload struct {
	UploadPath        string    `json:"upload_path"`        // Absolute path to the directory to delete on the Talis server (e.g., $TALIS_UPLOAD_DIR/[OWNER_ID]/[REQUEST_ID]/)
	DeletionTimestamp time.Time `json:"deletion_timestamp"` // Time after which deletion should occur
}
