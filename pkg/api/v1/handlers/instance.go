// Package handlers provides HTTP request handlers for the API
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/celestiaorg/talis/pkg/models"
	"github.com/celestiaorg/talis/pkg/types"
)

var (
	defaultUploadDir        = os.ExpandEnv("$HOME/talis/uploads")
	defaultDirCleanupWindow = 24 * time.Hour
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	*APIHandler
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(api *APIHandler) *InstanceHandler {
	return &InstanceHandler{
		APIHandler: api,
	}
}

// ListInstances handles the request to list all instances
func (h *InstanceHandler) ListInstances(c *fiber.Ctx) error {
	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Handle status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status, err := models.ParseInstanceStatus(statusStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(types.ErrInvalidInput(fmt.Sprintf("invalid instance status: %v", err)))
		}
		opts.InstanceStatus = &status
	} else if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		// By default, exclude terminated instances if not including deleted
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// TODO: should check for OwnerID and filter by it

	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(fmt.Sprintf("instance id is required: %v", err)))
	}

	if instanceID <= 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("instance id must be positive"))
	}

	// Get instance using the service
	instance, err := h.instance.GetInstance(c.Context(), models.AdminID, uint(instanceID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance: %v", err),
		})
	}

	return c.JSON(instance)
}

// CreateInstance handles the request to create instances, potentially including file uploads.
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	// --- 1. Parse Multipart Form which is now required for the upload ---
	_, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(fmt.Sprintf("failed to parse multipart form: %v", err)))
	}

	// --- 2. Unmarshal JSON data from form field ---
	requestDataJSON := c.FormValue("request_data")
	if requestDataJSON == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("missing 'request_data' field in form"))
	}

	var instanceReqs []types.InstanceRequest
	if err := json.Unmarshal([]byte(requestDataJSON), &instanceReqs); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(fmt.Sprintf("failed to unmarshal 'request_data': %v", err)))
	}

	if len(instanceReqs) == 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("at least one instance request is required in 'request_data'"))
	}

	// --- 3. Prepare for File Uploads ---
	// Files will be uploaded to a unique directory based on this request ID
	// Individual instance requests will have their own subdirectories within this request ID directory based on the OwnerID to allow for multiple users to upload files to the same request ID and avoid conflicts.
	// This allows for the cleanup of all files for a given request ID and all of the individual instance request files.
	requestID := uuid.NewString()
	uploadBaseDir := os.Getenv("TALIS_UPLOAD_DIR")
	if uploadBaseDir == "" {
		uploadBaseDir = defaultUploadDir
	}
	uniqueRequestDir := filepath.Join(uploadBaseDir, requestID)

	log.Printf("Creating unique request directory: %s", uniqueRequestDir)

	// --- Temporarily Commented Out During Rebase ---
	/*
		// Create a directory clean up task first and this protects against any failures during upload. Even if there are no uploads this is a low cost operation and is a no-op if the directory is not created.
		deletionTimestamp := time.Now().Add(defaultDirCleanupWindow) // Configurable?
		payload := types.UploadDeletionPayload{
			UploadPath:        uniqueRequestDir,
			DeletionTimestamp: deletionTimestamp,
		}
		payloadJSON, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			err = fmt.Errorf("failed to marshal cleanup task payload for request %s: %w", requestID, marshalErr)
			log.Print(err)
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer(err.Error()))
		}
		// Use the AdminIDs as the ownerID for the cleanup task to ensure it is always run
		cleanupTask := &models.Task{
			OwnerID:   models.AdminID,
			ProjectID: models.AdminProjectID,
			Name:      fmt.Sprintf("delete-upload-%s", requestID),
			Action:    models.TaskActionDeleteUpload,
			Status:    models.TaskStatusPending,
			Payload:   payloadJSON,
		}
		err = h.task.Create(c.Context(), cleanupTask)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer(err.Error()))
		}
		log.Printf("Prepared cleanup task for request %s, path: %s", requestID, uniqueRequestDir)
	*/
	// --- End Commented Out ---

	// --- 4. Process Instance Requests and Files ---
	for i := range instanceReqs {
		req := &instanceReqs[i] // Use pointer to modify original slice element
		req.Action = "create"   // Set default action
		ownerIDStr := strconv.FormatUint(uint64(req.OwnerID), 10)

		// Process Payload File
		userPayloadPath := req.PayloadPath
		if userPayloadPath != "" {
			// File not yet saved in this batch, attempt to save it
			serverPath, err := h.uploadFile(c, filepath.Join(uniqueRequestDir, ownerIDStr), userPayloadPath)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).
					JSON(types.ErrInvalidInput(fmt.Sprintf("failed to upload payload file '%s': %v", userPayloadPath, err)))
			}
			log.Printf("Saved payload file '%s' to '%s' for request %s", userPayloadPath, serverPath, requestID)
			// Update request with server-side path
			req.PayloadPath = serverPath
		}

		// --- Temporarily Commented Out During Rebase ---
		/*
			// Process Tar Archive File
			userTarPath := req.TarArchivePath
			if userTarPath != "" {
				serverPath, err := h.uploadFile(c, filepath.Join(uniqueRequestDir, ownerIDStr), userTarPath)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).
						JSON(types.ErrInvalidInput(fmt.Sprintf("failed to upload tar archive file '%s': %v", userTarPath, err)))
				}
				log.Printf("Saved tar archive '%s' to '%s' for request %s", userTarPath, serverPath, requestID)
				// Update request with server-side path
				req.TarArchivePath = serverPath
			}
		*/
		// --- End Commented Out ---

		// --- 6. Validate *after* paths are updated ---
		if err := req.Validate(); err != nil {
			// Assign the error to the named return variable for the deferred cleanup
			err = fmt.Errorf("validation failed for request %d: %w", i, err)
			return c.Status(fiber.StatusBadRequest).
				JSON(types.ErrInvalidInput(err.Error()))
		}
	}

	taskNames, err := h.instance.CreateInstance(c.Context(), instanceReqs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}
	log.Printf("Successfully processed instance creation request %s", requestID)
	return c.Status(fiber.StatusCreated).
		JSON(types.Success(
			ResponseWithTaskNames{
				TaskNames: taskNames,
			}))
}

// uploadFile uploads a file to the given directory and returns the server path. If a file already exists at the location it is a no-op and assumes this is expected due to multiple instances using the same uploaded file.
func (h *InstanceHandler) uploadFile(c *fiber.Ctx, dirPath, uploadFilePath string) (string, error) {
	// Get the file from the form data
	fileHeader, err := c.FormFile(uploadFilePath)
	if err != nil {
		return "", c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(fmt.Sprintf("upload file '%s' not found in form data: %v", uploadFilePath, err)))
	}

	// Create directory if it doesn't exist yet
	if err := os.MkdirAll(dirPath, 0750); err != nil {
		return "", c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(fmt.Sprintf("failed to create upload directory '%s': %v", dirPath, err)))
	}

	// Save the file to the directory
	targetServerPath := filepath.Join(dirPath, filepath.Base(fileHeader.Filename))
	// Check if file already exists
	if _, err := os.Stat(targetServerPath); err == nil {
		return targetServerPath, nil
	}
	if err := c.SaveFile(fileHeader, targetServerPath); err != nil {
		return "", c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(fmt.Sprintf("failed to save upload file '%s' to '%s': %v", uploadFilePath, targetServerPath, err)))
	}

	return targetServerPath, nil
}

// GetPublicIPs returns a list of public IPs for all instances
func (h *InstanceHandler) GetPublicIPs(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all public IPs...")

	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Only apply default status filter if IncludeDeleted is false
	if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// Get instances with their details using the service
	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		fmt.Printf("❌ Error getting public IPs: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get public IPs: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Extract the public IPs from the instances
	publicIPs := make([]types.PublicIPs, len(instances))
	for i, instance := range instances {
		publicIPs[i] = types.PublicIPs{
			PublicIP: instance.PublicIP,
		}
	}

	// Return instances with pagination info
	return c.JSON(types.PublicIPsResponse{
		PublicIPs: publicIPs,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// GetAllMetadata returns a list of all instance details
func (h *InstanceHandler) GetAllMetadata(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all instance metadata...")

	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Only apply default status filter if IncludeDeleted is false
	if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// Get instances with their details using the service
	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		fmt.Printf("❌ Error getting instance: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance metadata: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Return instances with pagination info
	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// GetInstances handles the request to list instances
func (h *InstanceHandler) GetInstances(c *fiber.Ctx) error {
	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Handle status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status, err := models.ParseInstanceStatus(statusStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid instance status: %v", err),
			})
		}
		opts.InstanceStatus = &status
	}

	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// TerminateInstances handles the request to terminate instances
func (h *InstanceHandler) TerminateInstances(c *fiber.Ctx) error {
	var deleteReq types.DeleteInstancesRequest
	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	if deleteReq.ProjectID == 0 || len(deleteReq.InstanceIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project id and instance ids are required",
		})
	}

	err := h.instance.Terminate(c.Context(), deleteReq.OwnerID, deleteReq.ProjectID, deleteReq.InstanceIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to terminate instances: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).
		JSON(types.Success(nil))
}
